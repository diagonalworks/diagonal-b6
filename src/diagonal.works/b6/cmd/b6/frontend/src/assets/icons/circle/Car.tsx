import type { SVGProps } from 'react';
const SvgCar = (props: SVGProps<SVGSVGElement>) => (
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
            d="m16.096 8.704-1.192-1.108-1.058-2.115A1.01 1.01 0 0 0 12.981 5H7.019a1.01 1.01 0 0 0-.865.48L5.096 7.597 3.904 8.704a.48.48 0 0 0-.154.353v4.116a.48.48 0 0 0 .48.48h1.924c.192 0 .48-.192.48-.384v-.577h6.731v.481c0 .192.193.48.385.48h2.02a.48.48 0 0 0 .48-.48V9.057a.48.48 0 0 0-.154-.353m-8.98-2.742h5.769l.961 1.923H6.154zm.48 4.423c0 .192-.288.384-.48.384h-2.02c-.192 0-.384-.288-.384-.48V9.23c.096-.289.288-.481.576-.385l1.924.385c.192 0 .384.288.384.48zm7.693-.096c0 .192-.193.48-.385.48h-2.02c-.192 0-.48-.192-.48-.384v-.673c0-.193.192-.481.385-.481l1.922-.385c.289-.096.481.096.578.385z"
        />
    </svg>
);
export default SvgCar;
