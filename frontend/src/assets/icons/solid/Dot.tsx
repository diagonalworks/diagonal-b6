import type { SVGProps } from 'react';
const SvgDot = (props: SVGProps<SVGSVGElement>) => (
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
            d="M8.5 14.25a4.25 4.25 0 1 0 0-8.5 4.25 4.25 0 0 0 0 8.5Z"
        />
    </svg>
);
export default SvgDot;
