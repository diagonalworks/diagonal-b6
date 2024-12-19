import { color } from "d3-color";
import { Color } from "deck.gl/typed";

/**
 * Convert an RGB color to a hex string.
 * @param r The red value.
 * @param g The green value.
 * @param b The blue value.
 * @returns The hex string.
 * @example
 * rgbToHex(100, 106, 195) = '#646ac3'
 */
export const rgbToHex = (r: number, g: number, b: number) =>
	"#" +
	[r, g, b]
		.map((x) => {
			const hex = x.toString(16);
			return hex.length === 1 ? "0" + hex : hex;
		})
		.join("");

/**
 * Convert a hex color to an RGB array.
 * @param c The hex color.
 * @param alpha The alpha value.
 * @returns The RGB array.
 * @example
 * hexToRgbArray('#646ac3') = [100, 106, 195, 255]
 */
export const colorToRgbArray = (c: string, alpha: number = 255) => {
	const rgbColor = color(c)?.rgb();
	if (!rgbColor) return [0, 0, 0, 0] as Color;
	return [rgbColor.r, rgbColor.g, rgbColor.b, alpha] as Color;
};

/**
 * Check if a string is a hex color string.
 * @param color The color string.
 * @returns Whether the string is a hex color string.
 * @example
 * isColorHex('#646ac3') = true
 * isColorHex('646ac3') = false
 */
export const isColorHex = (color: string) => {
	return /^#[0-9A-F]{6}$/i.test(color);
};
