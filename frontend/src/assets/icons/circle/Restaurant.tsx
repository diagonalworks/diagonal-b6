import type { SVGProps } from 'react';
const SvgRestaurant = (props: SVGProps<SVGSVGElement>) => (
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
            d="m7.045 3.75-.789 4.338c-.115.635 1.406.932 1.38 1.578l-.197 5.127c-.03.789.79.79.79.79s.818-.001.788-.79L8.82 9.666c-.025-.645 1.367-.931 1.38-1.578L9.412 3.75h-.395l.197 3.155-.591.395-.197-3.55H8.03L7.834 7.3l-.592-.395.197-3.155zm6.705 0c-.58 0-1.55.517-1.937 1.291-.322.58-.43 1.879-.43 2.653v1.972c0 .646.861.789 1.184.789l-.395 4.338c-.071.786.79.79.79.79s.788 0 .788-.79z"
        />
    </svg>
);
export default SvgRestaurant;
