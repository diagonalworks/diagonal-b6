import type { SVGProps } from 'react';
const SvgParking = (props: SVGProps<SVGSVGElement>) => (
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
            d="M4.5 3.75h-.25v11.5h2.5v-4H9a3.75 3.75 0 1 0 0-7.5zm2.25 5v-2.5H9a1.25 1.25 0 1 1 0 2.5z"
        />
    </svg>
);
export default SvgParking;
