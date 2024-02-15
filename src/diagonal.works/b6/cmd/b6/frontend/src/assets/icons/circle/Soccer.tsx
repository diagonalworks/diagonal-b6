import type { SVGProps } from 'react';
const SvgSoccer = (props: SVGProps<SVGSVGElement>) => (
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
            d="M13.032 4.004a1.504 1.504 0 1 1-3.008 0 1.504 1.504 0 0 1 3.008 0m0 9.525a1.003 1.003 0 1 0 0 2.005 1.003 1.003 0 0 0 0-2.005m1.845-4.923-1.915-1.915a.48.48 0 0 0-.371-.18H5.512a.501.501 0 0 0 0 1.002H8.22L5.011 13.83a.5.5 0 0 0 0 .2.512.512 0 0 0 1.003.21l1.002-1.714h2.006l-1.935 4.252a.5.5 0 0 0-.07.26.511.511 0 0 0 1.002.2l4.712-9.404 1.444 1.484a.501.501 0 0 0 .702-.712"
        />
    </svg>
);
export default SvgSoccer;
