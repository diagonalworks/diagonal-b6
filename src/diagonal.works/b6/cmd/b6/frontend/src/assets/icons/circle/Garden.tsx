import type { SVGProps } from 'react';
const SvgGarden = (props: SVGProps<SVGSVGElement>) => (
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
            d="M16.25 10.451c0 3.762-2.489 6.819-6.25 6.819s-6.25-3.057-6.25-6.819a6.06 6.06 0 0 1 5.682 4.103v-5.24H6.59A1.705 1.705 0 0 1 4.886 7.61V4.201a.568.568 0 0 1 1.023-.34l1.739 2.272 1.875-3.409a.568.568 0 0 1 .954 0l1.875 3.41 1.739-2.274a.568.568 0 0 1 1.023.341v3.41c0 .94-.764 1.704-1.705 1.704h-2.84v5.239a6.06 6.06 0 0 1 5.681-4.103"
        />
    </svg>
);
export default SvgGarden;
