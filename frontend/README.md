# b6-visual

b6-visual is a web application for visualization that enables interactwith the b6 engine.

### Development

The frontend is a [React](https://reactjs.org/) application written in [TypeScript](https://www.typescriptlang.org/). We use [Vite](https://vitejs.dev/) as the build tool.

#### Getting started

We use [pnpm](https://pnpm.io/) as the package manager, you can install it with the following command:

```bash
npm install -g pnpm
```

Then, you can start the development server with the following commands:

```bash
# Install dependencies
pnpm install

# Start the development server
pnpm dev
```

The development server will start at `http://localhost:5173`.

#### Storybook

We use [Storybook](https://storybook.js.org/) for component development. You can start the Storybook server with the following command:

```bash
pnpm storybook
```

The Storybook server will start at `http://localhost:6006`.
