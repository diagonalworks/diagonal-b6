import type { SVGProps } from 'react';
const SvgNaturalAreas = (props: SVGProps<SVGSVGElement>) => (
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
            d="M10 3.75c-.288 0-.385.192-.577.385l-5.577 9.134c-.096.096-.096.289-.096.385 0 .48.385.673.673.673h11.154c.384 0 .673-.192.673-.673 0-.193 0-.193-.096-.385l-5.48-9.134c-.193-.193-.385-.385-.674-.385m0 1.442 3.173 5.289h-.77l-1.442-1.443L10 10.481l-.962-1.443-1.442 1.443h-.865z"
        />
    </svg>
);
export default SvgNaturalAreas;
