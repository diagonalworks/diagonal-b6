import type { SVGProps } from 'react';
const SvgBakery = (props: SVGProps<SVGSVGElement>) => (
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
            d="m8.162 7.402.588 4.265a.816.816 0 0 0 .833.833h.834a.814.814 0 0 0 .833-.833l.588-4.265C11.838 6.25 10 6.25 10 6.25s-1.84 0-1.838 1.152m-1.495.515c-1.25 0-1.25.833-1.25.833l.833 3.333H7.5a.66.66 0 0 0 .662-.637L7.5 7.916zM5 10a1.28 1.28 0 0 0-.883.343c-.21.191-.34.453-.367.735v1.839h.735a.85.85 0 0 0 .932-.834zm8.333-2.083c1.25 0 1.25.833 1.25.833l-.833 3.333H12.5a.66.66 0 0 1-.662-.637l.662-3.53zM15 10c.327-.003.643.12.883.343.21.191.34.453.367.735v1.839h-.735a.85.85 0 0 1-.932-.834z"
        />
    </svg>
);
export default SvgBakery;
