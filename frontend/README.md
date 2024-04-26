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

We use [Jotai](https://jotai.org/) and [Immer](https://immerjs.github.io/immer/)for state management. Jotai takes an atomic approach to global React state management, allowing for a primitive and flexible way to manage state. Immer is used to update the state in an immutable way.

We defined a custom [atom with storage](https://jotai.org/docs/utilities/storage#atomwithstorage) to persist the view state in the URL search params. This allows us to share the current view with others.

We keep the global app state in an [atomWithImmer](https://jotai.org/docs/extensions/immer#withimmer), so that we can update the state in an immutable way, while having a single source of truth for the app state.

##### Components

We use Radix UI as the basis for the design system. Radix UI is a collection of low-level UI components that are unstyled by default. This allows us to have full control over the styling of the components.

Design system components are defined in `@/components/system`. These components are the presentational building blocks for the application and are used to create more complex components.

To connect these presentational components with the app state and logic, we define adapter components in `@/components/adapters`. These components are responsible for connecting the presentational components with the app state and logic.

Stack and Line adapters require their own state management, as such we use React Context with custom providers, these also make use of Immer for immutability.

##### API

We use [@tanstack/query](https://tanstack.com/query/latest) to fetch data from the API. This library provides hooks for fetching, caching and updating asynchronous data in React.

We currently define a proxy in `vite.config.ts` to forward API requests from `/api` to the backend.

##### Typing

We use [TypeScript](https://www.typescriptlang.org/) for static typing. We define custom types in the `@/types` directory. Some types are generated from the proto files using [ts-proto](https://ts-proto.readthedocs.io/).

## Note on fidelity

The version of the frontend in this PR is a simplified version of the final product, it intends to lay the architecture foundations and achieve the core functionality of Baseline. Therefore some features are not implemented or are implemented with a lower fidelity than the final product. Also, some design decisions may be subject to change in the future. Code quality, testing, and performance are not the main focus of this PR, but they will be addressed in future PRs.

#### Storybook

We use [Storybook](https://storybook.js.org/) for component development. You can start the Storybook server with the following command:

```bash
pnpm storybook
```

The Storybook server will start at `http://localhost:6006`.
