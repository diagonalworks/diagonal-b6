import type { SVGProps } from 'react';
const SvgSoccer = (props: SVGProps<SVGSVGElement>) => (
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
            d="m13.578 8.331-.006-.006-.006-.006-1.715-1.715a.68.68 0 0 0-.517-.244H4.962a.701.701 0 0 0 0 1.402H6.99l-2.703 5.322-.015.03-.007.032a.7.7 0 0 0-.009.23.711.711 0 0 0 1.39.286l.812-1.388h1.273l-1.575 3.46a.7.7 0 0 0-.091.327.71.71 0 0 0 1.389.29l4.07-8.124L12.58 9.31l.008.01.01.007a.701.701 0 0 0 .982-.996Zm-3.203-2.374a1.604 1.604 0 1 0 0-3.207 1.604 1.604 0 0 0 0 3.207Zm1.354 6.72a1.152 1.152 0 1 0 0 2.304 1.152 1.152 0 0 0 0-2.305Z"
        />
    </svg>
);
export default SvgSoccer;
