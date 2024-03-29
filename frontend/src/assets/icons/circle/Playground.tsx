import type { SVGProps } from 'react';
const SvgPlayground = (props: SVGProps<SVGSVGElement>) => (
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
            d="M5.044 4.79a1.356 1.356 0 1 1 2.623.683 1.356 1.356 0 0 1-2.623-.684m9.94 10.093a.904.904 0 0 1-.723 1.057.9.9 0 0 1-.967-.443l-1.518-3.045-1.31.262-.199.072h-.054v1.943l.56-.136h.127a.47.47 0 0 1 .171.904l-4.518.904a.5.5 0 0 1-.153 0 .47.47 0 0 1-.181-.904l3.75-.75v-1.96l-2.91.523a.905.905 0 0 1-1.093-.614h-.018l-.904-3.678a.9.9 0 0 1 0-.388.9.9 0 0 1 .723-.678l4.247-.47V3.75h.199v3.705h.054l.199-.018.361-.073h.19c.242.059.39.3.334.543a.44.44 0 0 1-.442.361l-.443.072h-.253v2.585h.054l.199-.055 1.635-.325a.9.9 0 0 1 .904.542l1.807 3.56q.105.105.172.236m-4.97-6.497-2.259.244.669 2.656 1.59-.307z"
        />
    </svg>
);
export default SvgPlayground;
