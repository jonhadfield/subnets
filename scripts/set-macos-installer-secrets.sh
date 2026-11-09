#!/bin/sh
# Export Developer ID Installer from the login keychain and set the GitHub
# Actions secrets the release workflow needs for subnets_macos.pkg.
#
#   ./scripts/set-macos-installer-secrets.sh
#
# Run in Terminal.app (not a headless agent). `security export -t identities`
# dumps every identity with no filter on modern macOS, so this uses a small
# Swift SecItemExport instead. Keychain may prompt Allow.
#
# Requires: swift, openssl, gh (auth'd to the repo with secret-write).
set -eu

repo="${SUBNETS_REPO:-jonhadfield/subnets}"
summary_match="${MACOS_INSTALLER_SUMMARY:-Developer ID Installer}"

need() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "need $1 on PATH" >&2
        exit 1
    }
}

need swift
need openssl
need gh
need base64

if ! gh auth status -h github.com >/dev/null 2>&1; then
    echo "gh is not logged in to github.com" >&2
    exit 1
fi

work="$(mktemp -d)"
# shellcheck disable=SC2064
trap 'rm -rf "$work"' EXIT INT TERM

pass="$(openssl rand -hex 24)"
printf '%s' "$pass" >"$work/pass"

cat >"$work/export.swift" <<'SWIFT'
import Foundation
import Security

let outPath = CommandLine.arguments[1]
let password = CommandLine.arguments[2]
let needle = CommandLine.arguments.count > 3 ? CommandLine.arguments[3] : "Developer ID Installer"

let query: [String: Any] = [
    kSecClass as String: kSecClassIdentity,
    kSecReturnRef as String: true,
    kSecMatchLimit as String: kSecMatchLimitAll,
]
var raw: CFTypeRef?
let status = SecItemCopyMatching(query as CFDictionary, &raw)
guard status == errSecSuccess, let items = raw as? [SecIdentity] else {
    fputs("SecItemCopyMatching failed: \(status)\n", stderr)
    exit(1)
}

var found: SecIdentity?
for identity in items {
    var cert: SecCertificate?
    SecIdentityCopyCertificate(identity, &cert)
    guard let cert else { continue }
    let summary = SecCertificateCopySubjectSummary(cert) as String? ?? ""
    if summary.contains(needle) {
        found = identity
        fputs("matched: \(summary)\n", stderr)
        break
    }
}
guard let identity = found else {
    fputs("no identity whose summary contains \(needle) (scanned \(items.count))\n", stderr)
    exit(1)
}

let passData = password as NSString
var keyParams = SecItemImportExportKeyParameters(
    version: UInt32(SEC_KEY_IMPORT_EXPORT_PARAMS_VERSION),
    flags: [],
    passphrase: Unmanaged.passUnretained(passData),
    alertTitle: nil,
    alertPrompt: nil,
    accessRef: nil,
    keyUsage: nil,
    keyAttributes: nil
)
var cfData: CFData?
let expStatus = SecItemExport(identity, .formatPKCS12, [], &keyParams, &cfData)
guard expStatus == errSecSuccess, let cfData else {
    fputs("SecItemExport failed: \(expStatus) — unlock the keychain and Allow access, then retry\n", stderr)
    exit(1)
}
try (cfData as Data).write(to: URL(fileURLWithPath: outPath), options: .atomic)
fputs("wrote \(outPath) (\((cfData as Data).count) bytes)\n", stderr)
SWIFT

echo "Exporting '$summary_match' via SecItemExport…"
echo "If macOS asks, choose Allow (or Always Allow)."
swift "$work/export.swift" "$work/installer.p12" "$pass" "$summary_match"

# Prove the bag imports as Developer ID Installer before uploading.
tmpkc="$work/verify.keychain-db"
security create-keychain -p verify "$tmpkc" >/dev/null
security unlock-keychain -p verify "$tmpkc"
if ! security import "$work/installer.p12" -k "$tmpkc" -P "$pass" -T /usr/bin/pkgbuild >/dev/null 2>"$work/import.err"; then
    echo "import verification failed:" >&2
    cat "$work/import.err" >&2
    exit 1
fi
if ! security find-identity -v -p basic "$tmpkc" | grep -q "Developer ID Installer"; then
    echo "exported PKCS#12 does not contain Developer ID Installer" >&2
    security find-identity -v -p basic "$tmpkc" >&2 || true
    exit 1
fi
security delete-keychain "$tmpkc" >/dev/null 2>&1 || true

base64 <"$work/installer.p12" | tr -d '\n' >"$work/installer.b64"

echo "Setting secrets on $repo…"
gh secret set MACOS_INSTALLER_P12 --repo "$repo" <"$work/installer.b64"
gh secret set MACOS_INSTALLER_PASSWORD --repo "$repo" <"$work/pass"

echo "Done. Next release tag will build, notarize, staple, and upload subnets_macos.pkg."
echo "Confirm with: gh secret list --repo $repo | grep MACOS_INSTALLER"
