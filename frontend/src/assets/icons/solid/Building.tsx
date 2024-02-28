import type { SVGProps } from 'react';
const SvgBuilding = (props: SVGProps<SVGSVGElement>) => (
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
            d="M4.5 4.75h-.25v10.278h4.944V12.36h2.167v2.667h1.389V4.75H4.5Zm3.306 8.889H5.639V12.36h2.167zm0-2.667H5.639V9.694h2.167zm0-2.666H5.639V7.028h2.167zm3.555 2.666H9.194V9.694h2.167zm0-2.666H9.194V7.028h2.167z"
        />
    </svg>
);
export default SvgBuilding;
