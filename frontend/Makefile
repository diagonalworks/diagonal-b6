# pnpm --config.confirmModulesPurge=false prevents pnpm from blocking on
# stdin in the continuous build.

all:
	pnpm install --config.confirmModulesPurge=false
	pnpm build
	pnpm build-storybook
