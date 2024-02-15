import type { SVGProps } from 'react';
const SvgBuilding = (props: SVGProps<SVGSVGElement>) => (
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
            d="M5 3.75v12.222h5.556V12.64h3.333v3.333H15V3.75zm4.444 11.111H6.111V12.64h3.333zm0-3.333H6.111V9.306h3.333zm0-3.334H6.111V5.972h3.333zm4.445 3.334h-3.333V9.306h3.333zm0-3.334h-3.333V5.972h3.333z"
        />
    </svg>
);
export default SvgBuilding;
