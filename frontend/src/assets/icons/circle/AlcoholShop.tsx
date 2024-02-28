import type { SVGProps } from 'react';
const SvgAlcoholShop = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={18}
        height={18}
        fill="none"
        viewBox="0 0 20 20"
        {...props}
    >
        <circle cx={10} cy={10} r={9.5} fill={props.fill} stroke="#fff" />
        <path
            fill="#fff"
            d="M15 6.827h-3.077v2.692c.002.7.476 1.31 1.154 1.485v2.746h-.385a.385.385 0 0 0 0 .77h1.539a.385.385 0 0 0 0-.77h-.385v-2.746A1.54 1.54 0 0 0 15 9.519V6.827m-.77 2.692a.77.77 0 0 1-1.538 0V7.596h1.539zM8.463 5.673v-.385a.385.385 0 0 0 0-.769v-.384a.385.385 0 0 0-.385-.385h-.77a.385.385 0 0 0-.384.385v.384a.385.385 0 0 0 0 .77v.384C6.923 6.773 5 8.035 5 9.135v4.615c0 .425.344.77.77.77h3.845a.846.846 0 0 0 .77-.77V9.135c0-1.039-1.923-2.423-1.923-3.462m-.77 7.308a1.923 1.923 0 1 1 0-3.846 1.923 1.923 0 0 1 0 3.846"
        />
    </svg>
);
export default SvgAlcoholShop;
