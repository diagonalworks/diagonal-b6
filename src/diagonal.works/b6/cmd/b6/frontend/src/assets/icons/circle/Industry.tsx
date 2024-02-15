import type { SVGProps } from 'react';
const SvgIndustry = (props: SVGProps<SVGSVGElement>) => (
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
            d="M15.938 3.75V15H3.75v-4.012a.47.47 0 0 1 .16-.347l2.812-3.02a.469.469 0 0 1 .778.357v2.813l2.963-3.16a.469.469 0 0 1 .787.347v5.147h2.813V3.75z"
        />
    </svg>
);
export default SvgIndustry;
