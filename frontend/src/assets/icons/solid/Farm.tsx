import type { SVGProps } from 'react';
const SvgFarm = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={20}
        height={20}
        fill="none"
        viewBox="0 0 17 20"
        {...props}
    >
        <path
            fill={props.fill}
            stroke="#fff"
            strokeWidth={0.5}
            d="m9.185 9.504-.03-.06-.055-.037-2.769-1.846-.139-.092-.138.092-2.77 1.846-.055.037-.03.06-.923 1.846-.026.053v4.001h2.346v-1.846h3.192v1.846h2.347v-4.002l-.027-.052zm5.315 5.9h.25v-9.02a1.635 1.635 0 0 0-3.27 0v9.02h3.02ZM5.52 11.21V9.865h1.345v1.346z"
        />
    </svg>
);
export default SvgFarm;
