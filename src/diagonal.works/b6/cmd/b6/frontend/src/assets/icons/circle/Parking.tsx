import type { SVGProps } from 'react';
const SvgParking = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={20}
        height={20}
        fill="none"
        viewBox="0 0 20 20"
        {...props}
    >
        <circle cx={10} cy={10} r={9.5} fill={props.fill} stroke="#fff" />
        <path
            fill="#fff"
            d="M6.25 3.75V17.5h2.5v-5h3.125a4.375 4.375 0 1 0 0-8.75zM8.75 10V6.25h3.125a1.875 1.875 0 0 1 0 3.75z"
        />
    </svg>
);
export default SvgParking;
