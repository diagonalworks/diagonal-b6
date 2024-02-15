import type { SVGProps } from 'react';
const SvgIndustry = (props: SVGProps<SVGSVGElement>) => (
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
            strokeLinejoin="round"
            strokeWidth={0.5}
            d="M14.5 3.75a.25.25 0 0 1 .25.25v11.077a.25.25 0 0 1-.25.25h-12a.25.25 0 0 1-.25-.25v-3.954a.71.71 0 0 1 .232-.517l2.761-2.964.018-.017a.71.71 0 0 1 1.004.064zm0 0h-1.846a.25.25 0 0 0-.25.25v8.98M14.5 3.75l-2.096 9.23M8.927 7.65l.012-.012a.711.711 0 0 1 1.196.526M8.927 7.65l.958.512m-.958-.512L6.442 10.3V8.163m2.485-.513-2.485.513m3.693 0h-.25m.25 0h-.25m.25 0v4.817m-.25-4.818v4.818h.25m0 0h2.269m-2.27 0v.25m0 0h2.27v-.25m-2.27.25L12.655 4v8.98h-.25m-2.27.25 2.27-.25M6.442 8.165a.7.7 0 0 0-.177-.475z"
        />
    </svg>
);
export default SvgIndustry;
