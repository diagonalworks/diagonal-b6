import type { SVGProps } from 'react';
const SvgHome = (props: SVGProps<SVGSVGElement>) => (
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
            d="M4.712 16.01c0 .133.107.24.24.24h3.605v-2.885h2.886v2.885h3.605a.24.24 0 0 0 .24-.24v-5.53H4.712zm11.473-6.896-.896-.787V4.712a.962.962 0 0 0-1.924 0v1.923l-3.19-2.798a.24.24 0 0 0-.34-.011l-.01.01-6.01 5.255a.24.24 0 0 0 .172.404l1.686.024H16.01a.24.24 0 0 0 .176-.405"
        />
    </svg>
);
export default SvgHome;
