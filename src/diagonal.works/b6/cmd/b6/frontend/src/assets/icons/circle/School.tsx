import type { SVGProps } from 'react';
const SvgSchool = (props: SVGProps<SVGSVGElement>) => (
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
            d="m8.117 6.294-2.342-.622.426-1.567a.486.486 0 0 1 .594-.338l1.404.377a.48.48 0 0 1 .341.589zm-4.346 7.073a.5.5 0 0 1-.004-.266l1.748-6.466 2.341.622-1.748 6.466a.5.5 0 0 1-.138.228l-1.4 1.295a.146.146 0 0 1-.238-.063zm8.632-5.772c-2.163 0-2.884-.72-2.884-2.884 2.163 0 2.884.721 2.884 2.884m-1.346 7.676c-1.317.202-3.005-1.64-3.384-3.653l.855-3.158a2.06 2.06 0 0 1 1.364-.512 2.6 2.6 0 0 1 1.504.47.93.93 0 0 0 1.047-.01c.406-.3.898-.462 1.403-.46.74 0 1.58.47 1.971 1.155 1.477 2.592-1.148 6.45-3.023 6.17a.8.8 0 0 1-.245-.09 1.37 1.37 0 0 0-1.252 0 .8.8 0 0 1-.24.088"
        />
    </svg>
);
export default SvgSchool;
