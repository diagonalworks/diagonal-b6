import type { SVGProps } from 'react';
const SvgBus = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={20}
        height={20}
        fill="none"
        viewBox="0 0 17 20"
        {...props}
    >
        <path
            fill={props.fill}
            stroke="#fff"
            strokeWidth={0.5}
            d="M3.5 13.09h-.25V5.819c0-1.138.93-2.068 2.068-2.068h6.364c1.138 0 2.068.93 2.068 2.068v7.273c0 .27-.069.49-.184.664a1.05 1.05 0 0 1-.409.355c-.113.057-.224.09-.316.11M3.5 13.09V5.819C3.5 4.818 4.318 4 5.318 4h6.364c1 0 1.818.818 1.818 1.818v7.273c0 .909-.91.909-.91.909m-9.09-.91-.25.001v.018a1 1 0 0 0 .018.165c.017.102.05.24.122.383s.183.294.355.409q.174.118.414.163M3.5 13.09l.66 1.138m8.68-.01V14h-.25m.25.22v.69c0 .637-.52 1.158-1.158 1.158s-1.16-.52-1.16-1.159v-.659H6.478v.66c0 .637-.52 1.158-1.159 1.158-.638 0-1.159-.52-1.159-1.159v-.68m8.682-.01a1.5 1.5 0 0 1-.232.03h-.011l-.004.001h-.002L12.59 14m0 0-8.432.229m.5-7.047a.2.2 0 0 1 .205-.205h7.272a.2.2 0 0 1 .205.205v2.727a.2.2 0 0 1-.205.205H4.864a.2.2 0 0 1-.205-.205zm0 5c0-.362.297-.66.66-.66.361 0 .658.298.658.66s-.297.659-.659.659a.66.66 0 0 1-.659-.66Zm6.364 0c0-.362.297-.66.659-.66s.659.298.659.66-.297.659-.66.659a.66.66 0 0 1-.658-.66Zm-5.25-6.614a.2.2 0 0 1-.205-.204.2.2 0 0 1 .205-.205h5.454a.2.2 0 0 1 .205.205.2.2 0 0 1-.205.204z"
        />
    </svg>
);
export default SvgBus;
