import type { SVGProps } from 'react';
const SvgPharmacy = (props: SVGProps<SVGSVGElement>) => (
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
            d="m12.611 11.82 1.472-1.487A3.1 3.1 0 0 0 15 8.125q0-1.305-.91-2.216Q13.182 5 11.876 5a3.09 3.09 0 0 0-2.208.917L8.18 7.389zM8.125 15a3.09 3.09 0 0 0 2.208-.917l1.486-1.472-4.43-4.43-1.472 1.486A3.1 3.1 0 0 0 5 11.875q0 1.305.91 2.215t2.215.91"
        />
    </svg>
);
export default SvgPharmacy;
