import type { SVGProps } from 'react';
const SvgFarm = (props: SVGProps<SVGSVGElement>) => (
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
            d="M10.48 8.558 7.597 6.635 4.712 8.558 3.75 10.48v3.846h1.923v-1.923H9.52v1.923h1.923V10.48zM8.559 10.48H6.635V8.558h1.923zm7.692 3.846h-2.885V5.192a1.442 1.442 0 1 1 2.885 0z"
        />
    </svg>
);
export default SvgFarm;
