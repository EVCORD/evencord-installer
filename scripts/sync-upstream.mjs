/*
 * Vencord, a Discord client mod
 * Copyright (c) 2026 Vendicated and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */
import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync, writeFileSync } from "node:fs";
const git = (...args) => execFileSync("git", args, { encoding: "utf8" }).trim();
git("fetch", "--no-tags", "https://github.com/Equicord/Equilotl.git", "main");
const upstream = git("rev-parse", "FETCH_HEAD");
if (spawnSync("git", ["merge-base", "--is-ancestor", upstream, "HEAD"]).status === 0) process.exit(0);
const merge = spawnSync("git", ["merge", "--no-ff", "--no-commit", upstream], { stdio: "inherit" });
if (merge.status) throw Error("Equilotl has source conflicts. Resolve them before publishing.");
const file = "buildinfo/git_data.go";
let source = readFileSync(file, "utf8");
const match = source.match(/InstallerTag\s*=\s*"(\d+)\.(\d+)\.(\d+)"/);
if (!match) throw Error("The installer version is unavailable.");
const version = `${match[1]}.${match[2]}.${Number(match[3]) + 1}`;
source = source.replace(match[0], `InstallerTag = "${version}"`);
writeFileSync(file, source);
git("add", file);
git("commit", "-m", `Integrate Equilotl ${upstream.slice(0, 9)} in ${version}`);
