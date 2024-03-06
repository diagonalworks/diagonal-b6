# pnpm --config.confirmModulesPurge=false prevents pnpm from blocking on
# stdin in the continuous build.

all:
	npm install -g pnpm
	pnpm install --config.confirmModulesPurge=false
	pnpm build
