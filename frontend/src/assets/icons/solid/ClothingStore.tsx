import type { SVGProps } from 'react';
const SvgClothingStore = (props: SVGProps<SVGSVGElement>) => (
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
            d="M5.27 3.75h-.091l-.07.058L2.34 6.116l-.09.075v2.674h2.77v5.539h6.96V8.865h2.77V6.191l-.09-.075-2.77-2.308-.069-.058h-1.629l-.07.138L8.5 7.133 6.877 3.888l-.069-.138H5.27Z"
        />
    </svg>
);
export default SvgClothingStore;
