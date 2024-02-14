import type { Preview } from '@storybook/react'
import {
  themes,
  ThemeProvider,
  ensure,
} from "@storybook/theming";

import "tailwindcss/tailwind.css";

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: '^on[A-Z].*' },
    controls: {
      matchers: {
       color: /(background|color)$/i,
       date: /Date$/i,
      },
    },
    docs: {
      theme: themes.light,
    }
  },
};

const withThemeProvider = (Story, context) => {
  const {
    parameters: { options = {}, docs = {} },
  } = context;

  let themeVars = docs.theme;
  
  if (!themeVars && options.theme) {
    themeVars = options.theme;
  }

  const theme = ensure(themeVars);
  return (
    <ThemeProvider theme={theme}>
      <Story {...context} />
    </ThemeProvider>
  )
};

export const decorators = [withThemeProvider];

export default preview;