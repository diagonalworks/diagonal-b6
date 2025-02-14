import { ArgumentParser } from "argparse";
import fs from "fs";
import defaultConfig from "tailwindcss/defaultConfig";
import resolveConfig from "tailwindcss/resolveConfig";

import { $IntentionalAny } from "@/utils/defs";

import { figma } from "./api";

const config = resolveConfig(defaultConfig);

const BASE_FONT_SIZE = 16;
const rem = (px: number) => `${((px / BASE_FONT_SIZE) * 1000) / 1000}rem`;

const unicaFontWeights = [
	{ fontWeight: 300, variant: "light" },
	{ fontWeight: 400, variant: "regular" },
	{ fontWeight: 500, variant: "medium" },
	{ fontWeight: 700, variant: "bold" },
];

const getFontSizeStyle = (fontSize: number) => {
	const fontSizeConfig = config.theme.fontSize;
	const fontSizeKey = Object.entries(fontSizeConfig).find(
		([, value]) => value[0] === rem(fontSize),
	)?.[0];

	return fontSizeKey ? `text-${fontSizeKey}` : `text-base`;
};

const main = async () => {
	const parser = new ArgumentParser({
		description: "Write typography layers from Figma to a .css file.",
	});
	parser.add_argument("-f", "--file", {
		help: "Figma file id",
		required: true,
	});
	parser.add_argument("-o", "--output", {
		help: "Output file",
		required: true,
	});

	parser.add_argument("-r", "--raw", {
		help: "file to output raw text styles from Figma",
	});

	const { file, output, raw } = parser.parse_args();
	const api = figma();

	const styles = await api.styles(file);
	const textStyles = Object.values(styles.nodes).filter(
		(n: $IntentionalAny) => n.document.type === "TEXT",
	);

	const typography = textStyles.map((n: $IntentionalAny) => {
		const { document } = n;
		return {
			name: document.name,
			fontFamily: document.style.fontFamily,
			fontWeight: document.style.fontWeight,
			fontSize: document.style.fontSize,
			lineHeight: document.style.lineHeightPx,
		};
	});

	if (raw) {
		fs.writeFileSync(raw, JSON.stringify(typography, null, 2));
	}

	const body = typography.find((t: $IntentionalAny) => t.name === "Body base");

	const title = typography.find(
		(t: $IntentionalAny) => t.name === "Title base",
	);

	const css = `
/* This file is generated by scripts/figma/typography.ts. Please don't edit directly */

@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
    ${unicaFontWeights
			.map(
				(w) => `
        @font-face {
            font-family: 'Unica 77';
            font-weight: ${w.fontWeight};
            src: url('https://diagonal.works/fonts/unica77-${w.variant}.woff') format('woff');
        }
        @font-face {
            font-family: 'Unica 77';
            font-weight: ${w.fontWeight};
            src: url('https://diagonal.works/fonts/unica77-${w.variant}.woff2') format('woff2');
        }`,
			)
			.join("\r\n")}
}

@layer components {
    .base {
    @apply ${getFontSizeStyle(body?.fontSize ?? 14)} text-graphite-100;
    }
    .title {
    @apply ${getFontSizeStyle(title?.fontSize ?? 16)} text-graphite-100;
    }
}`;

	fs.writeFileSync(output, css);
};

main().catch((error) => {
	console.error(error);
	process.exit(1);
});
