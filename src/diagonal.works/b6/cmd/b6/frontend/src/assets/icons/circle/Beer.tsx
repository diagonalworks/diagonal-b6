import type { SVGProps } from 'react';
const SvgBeer = (props: SVGProps<SVGSVGElement>) => (
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
            d="M14.167 7.269v-2.64s-.88-.879-3.959-.879c-3.078 0-3.958.88-3.958.88v2.639a8.15 8.15 0 0 0 .88 3.518 4.95 4.95 0 0 1 0 3.958s0 .88 3.078.88c3.079 0 3.079-.88 3.079-.88a4.95 4.95 0 0 1 0-3.958 8.15 8.15 0 0 0 .88-3.518m-3.959 7.476c-.7.031-1.402-.052-2.076-.246a7 7 0 0 0 .264-1.953h3.624c-.003.66.086 1.318.264 1.953a6.5 6.5 0 0 1-2.076.246m0-7.476a9.4 9.4 0 0 1-3.078-.44v-1.76a9.5 9.5 0 0 1 3.078-.44 9.5 9.5 0 0 1 3.079.44v1.76a9.4 9.4 0 0 1-3.079.44"
        />
    </svg>
);
export default SvgBeer;
