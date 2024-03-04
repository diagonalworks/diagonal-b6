import type { SVGProps } from 'react';
const SvgSwimming = (props: SVGProps<SVGSVGElement>) => (
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
            d="M12.176 5c-.094 0-.363.122-.363.122l-2.768 1.4c-.37.147-.515.735-.293 1.028l.809 1.174-3.31 1.691 1.666 1.25 2.085-1.25 2.083 1.25.835-.835-2.5-3.333 2.13-1.275c.441-.222.37-.587.37-.809-.003-.175-.3-.413-.744-.413m1.784 2.5a1.459 1.459 0 1 0-.002 2.918A1.459 1.459 0 0 0 13.96 7.5m-8.127 4.167-2.083 1.25v1.25l2.083-1.25 2.084 1.25 2.085-1.25 2.083 1.25 1.665-1.25 2.5 1.25v-1.25l-2.5-1.25-1.665 1.25-2.083-1.25-2.085 1.25z"
        />
    </svg>
);
export default SvgSwimming;
