import type { SVGProps } from 'react';
const SvgReligiousChristian = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={20}
        height={20}
        fill="none"
        viewBox="0 0 16 20"
        {...props}
    >
        <path
            fill={props.fill}
            stroke="#fff"
            strokeLinejoin="round"
            strokeWidth={0.5}
            d="M8.464 3h-.918m.918 0s.87 0 .87.889v2.417h.25M8.463 3v-.25h-.918V3m.918 0v-.25h.019a1 1 0 0 1 .16.017 1.3 1.3 0 0 1 .37.122c.138.07.284.18.394.35s.176.385.176.65v2.417M7.546 3 12 6.556H9.583v-.25M7.546 3l2.037 6.472M7.546 3v-.25h-.018a1 1 0 0 0-.16.016 1.4 1.4 0 0 0-.37.115 1 1 0 0 0-.4.335 1.1 1.1 0 0 0-.181.633v2.457H4a.25.25 0 0 0-.25.25v2.666c0 .138.112.25.25.25h2.417M7.546 3l-1.13 6.472m3.167-3.166H12a.25.25 0 0 1 .25.25v2.666a.25.25 0 0 1-.25.25H9.583m0 0v6.861a.25.25 0 0 1-.25.25H6.667a.25.25 0 0 1-.25-.25v-6.86m3.166 0H6.417"
        />
    </svg>
);
export default SvgReligiousChristian;
