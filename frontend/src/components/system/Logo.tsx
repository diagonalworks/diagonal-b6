import type { SVGProps } from 'react';

const Logo = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={24}
        height={24}
        fill="none"
        viewBox="0 0 32 21"
        {...props}
    >
        <g fill="currentColor">
            <path d="m15.56 13.262-4.54-.874-4.594 8.385-2.631-1.442 4.104-7.49L0 10.525l.493-2.96 10.002 1.667 2.176.419-6.75-7.403L8.139.228l7.423 8.14z" />
            <path
                fillRule="evenodd"
                d="M16.026 12.334v-2.121l.733.732 12.027 1.716a1.75 1.75 0 1 1-.206 1.486l-10.054-1.435 3.41 3.41a1.75 1.75 0 1 1-1.062 1.06z"
                clipRule="evenodd"
            />
        </g>
    </svg>
);
export default Logo;
