import type { SVGProps } from 'react';
const SvgLandmarks = (props: SVGProps<SVGSVGElement>) => (
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
            d="M14.546 13.75h-.455v-.454c0-.273-.182-.455-.455-.455h-.454V8.295h.909L15 6.477q-1.364.137-2.727 0a8.6 8.6 0 0 1-1.819-1.818v-.454c0-.273-.181-.455-.454-.455s-.455.182-.455.455v.454a8.6 8.6 0 0 1-1.818 1.818q-1.364.137-2.727 0l.91 1.818h.908v4.546h-.454c-.273 0-.455.182-.455.455v.454h-.454c-.273 0-.455.182-.455.454v.455h10v-.455c0-.272-.182-.454-.454-.454m-5-.91H7.726V8.296h1.818zm2.727 0h-1.819V8.296h1.819z"
        />
    </svg>
);
export default SvgLandmarks;
