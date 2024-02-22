import type { SVGProps } from 'react';
const SvgClothingStore = (props: SVGProps<SVGSVGElement>) => (
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
            d="M6.635 5 3.75 7.404v2.404h2.885v5.769h6.73v-5.77h2.885V7.405L13.365 5h-1.442L10 8.846 8.077 5z"
        />
    </svg>
);
export default SvgClothingStore;
