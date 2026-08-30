import { build, context } from "esbuild";
import { copyFileSync, cpSync, mkdirSync, rmSync } from "node:fs";

const watch = process.argv.includes("--watch");
const dist = new URL("./dist/", import.meta.url).pathname;

rmSync(dist, { recursive: true, force: true });
mkdirSync(dist, { recursive: true });

const options = {
  entryPoints: {
    background: "src/background.ts",
    content: "src/content.ts",
    viewer: "src/viewer.ts",
  },
  bundle: true,
  outdir: dist,
  format: /** @type {"iife"} */ ("iife"),
  target: "chrome109",
};

if (watch) {
  const ctx = await context(options);
  await ctx.watch();
} else {
  await build(options);
}

// Copy static assets
copyFileSync("manifest.json", `${dist}/manifest.json`);
copyFileSync("src/viewer.html", `${dist}/viewer.html`);
cpSync("icons", `${dist}/icons`, { recursive: true });

console.log(`${watch ? "watching" : "built"} -> dist/`);
