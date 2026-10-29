# Installer identity and compatibility improvements

The installer uses the shared #191A1E background, #34353B surfaces, #805263 accent and #F0EDEB text. GUI styling reads the embedded contract; PNG/ICO artwork and Windows resources match the core. The editable artwork generator remains in the core repository.

Implemented fixes:

- Existing installations reject releases whose signed compatibility list omits the installed original Discord archive.
- Activation checks the actual staged bytes again immediately before writing the immutable payload.
- A stale cached artifact is refreshed instead of being reused for a different release. Successful installer-owned staging is cleaned with a constrained path check.
- Failed GUI discovery clears the previous target; release checks reject rollback and reused identities instead of reporting an outdated release as current.
- Self-update proof metadata is written to a temporary file, flushed, then renamed. A failed proof write no longer truncates the previous proof. Verified side-by-side executable staging remains unchanged.
- Shared host persistence requires a newer Discord version, fresh signed compatibility metadata, matching EvenCord data roots and unlinked ownership files. Production host repatching remains off.

Focused fixture checks cover existing-install compatibility refusal, fresh/repair/uninstall preservation, and self-update failure plus complete/repeated proof staging. Windows CLI, local-preview CLI and GUI binaries compile. No installer or Discord runtime operation was performed. Production trust is empty, capabilities remain disabled, Actions remain disabled and no releases are published.

See the core [ecosystem record](https://github.com/alrimii/evencord/blob/main/ECOSYSTEM.md) for official plugin discovery, Reclip and Quoter Enhanced. The installed working preview is unchanged; a subsequent local installation requires its own explicit action.
