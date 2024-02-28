import type { SVGProps } from 'react';
const SvgGarden = (props: SVGProps<SVGSVGElement>) => (
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
            d="M8.747 4.073a.455.455 0 0 0-.629.135L6.75 6.696zm0 0q.082.053.135.135zm5.003 6.317v-.252l-.252.002a5.1 5.1 0 0 0-4.293 2.419V9.73h2.022c.891 0 1.614-.723 1.614-1.614V5.39a.705.705 0 0 0-1.268-.423l-1.157 1.512-1.315-2.392-.004-.007-.005-.008a.705.705 0 0 0-1.184 0l-.005.008-.004.007L6.584 6.48 5.426 4.966a.704.704 0 0 0-1.267.424v2.727c0 .891.723 1.614 1.614 1.614h2.022v2.828a5.1 5.1 0 0 0-4.293-2.42l-.252-.001v.252c0 3.117 2.075 5.704 5.25 5.704s5.25-2.587 5.25-5.704Z"
        />
    </svg>
);
export default SvgGarden;
