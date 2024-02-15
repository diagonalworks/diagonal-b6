import type { SVGProps } from 'react';
const SvgShop = (props: SVGProps<SVGSVGElement>) => (
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
            d="M15.763 7.7h-1.807l-.386-2.3a1.98 1.98 0 0 0-1.392-1.472 3.7 3.7 0 0 0-1.067-.178H8.899a3.7 3.7 0 0 0-1.067.178c-.7.197-1.234.762-1.393 1.471l-.385 2.302H4.247a.494.494 0 0 0-.474.642l1.857 6.222a1.48 1.48 0 0 0 1.412 1.037h5.926a1.48 1.48 0 0 0 1.402-1.037l1.857-6.222a.494.494 0 0 0-.464-.642m-8.701 0 .355-2.143a.9.9 0 0 1 .731-.691c.243-.077.496-.12.75-.128h2.213q.389.01.76.128c.362.06.652.333.731.691l.346 2.144H7.022z"
        />
    </svg>
);
export default SvgShop;
