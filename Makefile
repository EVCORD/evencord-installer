.PHONY: all clean build help

PLATFORM 	?= $(shell go env GOOS)
ARCH		?= $(shell go env GOARCH)

VERSION 	?= 0.1.0-dev.0
HASH 		:= $(shell git rev-parse --short HEAD | xargs)

# "Developer ID Application: <name> (<team-id>)" or "-"
IDENTITY 	?= -

GUI 		?= 0
DMG 		?= 0
WAYLAND 	?= 0

LDFLAGS 	?= -s -w -X 'equilotl/buildinfo.InstallerGitHash=$(HASH)' -X 'equilotl/buildinfo.InstallerTag=$(VERSION)'
TAGS    	?= static
CGO 		:= 0
POSTFIX 	:=
WINARCH 	:=

ifeq ($(PLATFORM),windows)
ifeq ($(ARCH),arm64)
WINARCH 	:= -arm64
endif # ARCH
ifeq ($(GUI),0)
LDFLAGS 	+= -extldflags=-static
else
LDFLAGS 	+= -H=windowsgui -extldflags=-static
endif # GUI
endif # PLATFORM

ifeq ($(GUI),0)
TAGS 		+= cli
POSTFIX 	:= Cli
else
CGO 		:= 1
ifeq ($(WAYLAND),1)
TAGS 		+= wayland
endif # WAYLAND
endif # GUI

# macOS specific
MACOS_ARCHS := $(ARCH)
ifeq ($(ARCH),universal)
MACOS_ARCHS := amd64 arm64
endif

ifeq ($(PLATFORM),windows)
ifneq ($(ARCH),amd64)
$(error EvenCord Phase 2 supports Windows x64 development builds only)
endif
else
$(error EvenCord Phase 2 supports Windows x64 development builds only)
endif

all: build

build:
	mkdir -p build

ifeq ($(PLATFORM),darwin)
	for arch in $(MACOS_ARCHS); do \
		MACOSX_DEPLOYMENT_TARGET=10.13 \
		CGO_ENABLED=$(CGO) \
		GOOS=$(PLATFORM) \
		GOARCH=$$arch \
		go build \
			-v \
			-tags "$(TAGS)" \
			-ldflags "$(LDFLAGS)" \
			-o build/installer-$$arch; \
	done
ifeq ($(ARCH),universal)
	lipo -create \
		build/installer-amd64 \
		build/installer-arm64 \
		-output build/installer-universal
	rm build/installer-amd64 build/installer-arm64
endif # ARCH
ifeq ($(GUI),1)
	cp -R macos/Equilotl.app build/Equilotl.app
	/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $(VERSION)" build/Equilotl.app/Contents/Info.plist
	/usr/libexec/PlistBuddy -c "Set :CFBundleVersion $(HASH)" build/Equilotl.app/Contents/Info.plist
ifeq ($(ARCH),universal)
	mv build/installer-universal build/Equilotl.app/Contents/MacOS/Equilotl
else
	mv build/installer-$(ARCH) build/Equilotl.app/Contents/MacOS/Equilotl
endif # ARCH
	codesign --deep --force --options runtime --sign "$(IDENTITY)" build/Equilotl.app
ifeq ($(DMG),1)
ifeq ($(shell command -v create-dmg 2>/dev/null),)
	$(error create-dmg is not installed)
endif
	create-dmg \
		--volname "Equilotl" \
		--volicon macos/icon.icns \
		--background "macos/background.png" \
		--window-pos 200 120 \
		--window-size 510 340 \
		--icon-size 100 \
		--icon Equilotl.app 160 155 \
		--hide-extension Equilotl.app \
		--app-drop-link 350 155 \
		build/Equilotl.dmg \
		build/Equilotl.app
endif # DMG
else # GUI
	mv build/installer-$(ARCH) build/EquilotlCli-$(PLATFORM)
endif # GUI

else ifeq ($(PLATFORM),windows) # PLATFORM
	go run ./cmd/brand-resources $(if $(filter 0,$(GUI)),-cli,)
	go-winres make --in build/winres.json --arch $(ARCH) --product-version "0.1.0.0"
	CGO_ENABLED=$(CGO) GOOS=$(PLATFORM) GOARCH=$(ARCH) \
		go build -v -tags "$(TAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o build/EvenCordInstaller$(POSTFIX)-x64.exe

else # PLATFORM
	CGO_ENABLED=$(CGO) GOOS=$(PLATFORM) GOARCH=$(ARCH) go build \
		-v \
		-tags "$(TAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o build/Equilotl$(POSTFIX)-$(PLATFORM)
	chmod +x build/Equilotl$(POSTFIX)-$(PLATFORM)
endif

clean:
	rm -rf build

help:
	@printf '%s\n' \
		'EvenCord Phase 2: Windows x64 development builds only.' \
		'make PLATFORM=windows ARCH=amd64 GUI=0' \
		'make PLATFORM=windows ARCH=amd64 GUI=1' \
		'Go, go-winres and a compatible CGO compiler (GUI) are required.' \
		'VERSION defaults to 0.1.0-dev.0. Installation and distribution are disabled.'
