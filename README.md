# EvenCord Installer

Based on Equicord's Equilotl installer, with EvenCord identity and download repositories. The normal Windows, macOS and Linux discovery and patching paths are inherited from upstream. Platform artifacts are built by the release matrix; device runtime checks must be reported separately.

Download from [EvenCord Installer releases](https://github.com/alrimii/evencord-installer/releases/latest). Use Install, Repair or Uninstall for your Discord installation. Settings are retained. Close Discord and save clips before operating.

Version 1.1.0 migrates owned legacy managed Windows installations through their existing recovery engine before applying the normal loader. New installs use a stable EvenCord payload path. Core downloads are checked against their exact repository URL, published size and GitHub SHA-256 digest. The former Windows-only archive catalog no longer limits normal installs. Legacy records and original Discord archives are preserved during recovery.

Windows binaries are not Authenticode signed. Installer self-update is disabled; download new installer releases from GitHub. Client updates are delivered by EvenCord core independently.

Licensed under GPL. Original Equilotl and Vencord installer attribution is retained. EvenCord support is maintained in its own repository.
