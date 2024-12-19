import { create } from "@storybook/theming/create";

export default create({
	base: "dark",
	brandTitle: "b6 design system",
	brandUrl: "https://www.diagonal.works",
	brandTarget: "_self",
	colorPrimary: "#3b3c67",
	colorSecondary: "#646ac3",

	// UI
	appBg: "#3b3c67",
	appContentBg: "#fff",
	appPreviewBg: "#fff",
	appBorderColor: "#eeecff",
	appBorderRadius: 2,

	// Text colors
	textColor: "#f9f9fe",
	textInverseColor: "#25253c",
	textMutedColor: "#b3a7da",

	// Toolbar default and active colors
	barTextColor: "#646ac3",
	barSelectedColor: "#3b3c67",
	barBg: "#f9f9fe",
});
