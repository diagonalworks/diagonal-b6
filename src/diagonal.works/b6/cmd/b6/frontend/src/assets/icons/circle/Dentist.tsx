import type { SVGProps } from 'react';
const SvgDentist = (props: SVGProps<SVGSVGElement>) => (
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
            d="M7.136 15.922c-.909 0-.51-2.426-.782-4.543-.09-.69-.908-1.354-1.017-1.872-.346-1.762-.9-4.997 1.163-5.396 2.062-.4 2.125 1.817 3.525 1.817s1.426-2.144 3.497-1.817c2.072.327 1.445 3.543 1.172 5.36-.09.409-.999 1.345-1.072 1.872-.3 2.18.291 4.542-.726 4.542-.845 0-1.2-2.471-1.818-4.088-.245-.754-.581-1.327-1.053-1.327-1.4 0-1.627 5.452-2.89 5.452"
        />
    </svg>
);
export default SvgDentist;
