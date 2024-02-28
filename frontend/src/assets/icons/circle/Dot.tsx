import type { SVGProps } from 'react';
const SvgDot = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={18}
        height={18}
        fill="none"
        viewBox="0 0 20 20"
        {...props}
    >
        <path
            fill={props.fill}
            stroke="#fff"
            d="M13.5 10a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0Z"
        />
    </svg>
);
export default SvgDot;
