import type { SVGProps } from 'react';
const SvgHospital = (props: SVGProps<SVGSVGElement>) => (
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
            d="M9.423 2.5c-.692 0-1.154.462-1.154 1.154v4.615H3.654c-.692 0-1.154.462-1.154 1.154v1.154c0 .692.462 1.154 1.154 1.154h4.615v4.615c0 .692.462 1.154 1.154 1.154h1.154c.692 0 1.154-.462 1.154-1.154v-4.615h4.615c.692 0 1.154-.462 1.154-1.154V9.423c0-.692-.462-1.154-1.154-1.154h-4.615V3.654c0-.692-.462-1.154-1.154-1.154z"
        />
    </svg>
);
export default SvgHospital;
