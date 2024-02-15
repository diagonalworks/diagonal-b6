import type { SVGProps } from 'react';
const SvgCarehome = (props: SVGProps<SVGSVGElement>) => (
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
            fillRule="evenodd"
            d="M14.837 14.432A.56.56 0 0 0 15 14.04V8.2a.56.56 0 0 0-.214-.439L10.34 4.305a.555.555 0 0 0-.682 0L5.214 7.76a.56.56 0 0 0-.214.44v5.839a.556.556 0 0 0 .556.555h8.888a.56.56 0 0 0 .393-.163M9.135 6.827c0-.346.23-.577.577-.577h.577c.346 0 .576.23.576.577v2.308h2.308c.346 0 .577.23.577.577v.577c0 .346-.23.576-.577.576h-2.308v2.308c0 .346-.23.577-.576.577h-.577c-.347 0-.577-.23-.577-.577v-2.308H6.827c-.346 0-.577-.23-.577-.576v-.577c0-.347.23-.577.577-.577h2.308z"
            clipRule="evenodd"
        />
    </svg>
);
export default SvgCarehome;
