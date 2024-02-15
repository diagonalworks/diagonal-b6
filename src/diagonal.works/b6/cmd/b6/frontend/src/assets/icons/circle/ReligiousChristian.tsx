import type { SVGProps } from 'react';
const SvgReligiousChristian = (props: SVGProps<SVGSVGElement>) => (
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
            d="M8.333 3.455V6.5H5v3h3.333v8h3.334v-8H15v-3h-3.333v-3c0-1-1.087-1-1.087-1H9.432s-1.099 0-1.099.955"
        />
    </svg>
);
export default SvgReligiousChristian;
