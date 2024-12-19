export const ConditionalWrap = ({
	condition,
	wrap,
	children,
}: {
	condition: boolean;
	wrap: (children: JSX.Element) => JSX.Element;
	children: JSX.Element;
}) => (condition ? wrap(children) : children);
