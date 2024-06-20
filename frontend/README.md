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

#### Architecture

##### Map

For rendering the map we use [react-map-gl](https://visgl.github.io/react-map-gl/). We chose this library as it is a wrapper around [maplibre-gl-js](https://maplibre.org/maplibre-gl-js/docs/) and provides a react-friendly API, allowing also for WebGL rendering. It also integrates well with [deck.gl](https://deck.gl/), which enables us to render more complex visualizations on top of the map in the future.

For map styling, we follow the [MapLibre Style Specification](https://maplibre.org/maplibre-style-spec/transition/). The style is defined in the `diagonal-map-style.json`. This style can be edited using [Maputnik](https://github.com/maplibre/maputnik).

##### State management

We use [Zustand](https://docs.pmnd.rs/zustand/getting-started/introduction) and [Immer](https://immerjs.github.io/immer/)for state management. Zustand is a small and fast state management library that uses React hooks. Immer is used to create immutable state updates in a more readable way. We define multiple stores in the `@/stores` directory, each store is responsible for managing a specific part of the application state. For the core stores, the `@/stores/outliners` and `@/stores/map` depend on the `@/stores/world` store.

##### Components

We use Radix UI as the basis for the design system. Radix UI is a collection of low-level UI components that are unstyled by default. This allows us to have full control over the styling of the components.

Design system components are defined in `@/components/system`. These components are the presentational building blocks for the application and are used to create more complex components.

To connect these presentational components with the app state and logic, we define adapter components in `@/components/adapters`. These components are responsible for connecting the presentational components with the app state and logic.

Stack and Line adapters require their own state management, as such we use React Context with custom providers, these also make use of Immer for immutability.

##### API

We use AXIOS for making API requests. We define an API client in `@/api/client.ts` that is responsible for making requests to the API.
We use [@tanstack/query](https://tanstack.com/query/latest) for data fetching and caching.

##### Typing

We use [TypeScript](https://www.typescriptlang.org/) for static typing. We define custom types in the `@/types` directory. Some types are generated from the proto files using [ts-proto](https://ts-proto.readthedocs.io/).

#### Testing

// TODO

## Note on fidelity

As we're currently migrating the Baseline UI to React, we're focusing on achieving the core functionality of Baseline. Therefore, some features of the original Baseline UI are not yet implemented or are implemented with a lower fidelity than the final product. Some design decisions may be subject to change in the future.

#### Storybook

We use [Storybook](https://storybook.js.org/) for component development. You can start the Storybook server with the following command:

```bash
pnpm storybook
```

The Storybook server will start at `http://localhost:6006`.
