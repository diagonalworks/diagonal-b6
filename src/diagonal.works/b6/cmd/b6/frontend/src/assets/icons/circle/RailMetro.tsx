import type { SVGProps } from 'react';
const SvgRailMetro = (props: SVGProps<SVGSVGElement>) => (
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
            d="M8.333 3.75s-.625 0-.833.833L6.25 9.167v2.916c0 .834.833.834.833.834h5.834s.833 0 .833-.834V9.167L12.5 4.583c-.227-.833-.833-.833-.833-.833zM9.167 5h1.666s.447 0 .625.833l.625 2.917c.18.835-.833.833-.833.833h-2.5s-1.012.002-.833-.833l.625-2.917C8.72 5 9.167 5 9.167 5m-1.25 5.417a.833.833 0 1 1 0 1.666.833.833 0 0 1 0-1.666m1.458 0h1.25a.208.208 0 1 1 0 .416h-1.25a.208.208 0 1 1 0-.416m2.708 0a.833.833 0 1 1 0 1.666.833.833 0 0 1 0-1.666M7.188 13.75l-.938 2.5H7.5l.313-.833h4.375l.312.833h1.25l-.937-2.5h-1.25l.312.833h-3.75l.313-.833z"
        />
    </svg>
);
export default SvgRailMetro;
