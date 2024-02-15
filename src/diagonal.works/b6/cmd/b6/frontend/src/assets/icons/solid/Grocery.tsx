import type { SVGProps } from 'react';
const SvgGrocery = (props: SVGProps<SVGSVGElement>) => (
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
            strokeWidth={0.5}
            d="M10.974 6.094H2.166l.093.32 1.307 4.513.006.02.009.019q.007.015.013.024c.255.68.889 1.173 1.647 1.22l.02.009h5.888l-.013.076c-.02.099-.06.241-.142.406-.1.201-.144.275-.22.326-.075.05-.222.098-.587.098H4.095a.93.93 0 0 0-.714.304.98.98 0 0 0-.24.65c0 .228.077.465.24.649a.93.93 0 0 0 .714.303h.223a1.24 1.24 0 0 0-.193.688c0 .338.114.639.33.856.218.217.52.331.857.331.338 0 .64-.114.857-.33.217-.218.33-.52.33-.857a1.24 1.24 0 0 0-.193-.688h2.7a1.24 1.24 0 0 0-.194.688c0 .338.114.639.331.856s.518.331.857.331.639-.114.856-.33c.217-.218.331-.52.331-.857 0-.31-.095-.587-.277-.799.813-.228 1.416-.778 1.663-1.353a3.9 3.9 0 0 0 .297-1.2l.004-.082.001-.023v-.01l-.25-.001h.25V6.344c0-.214.048-.398.124-.52.071-.112.16-.167.282-.167h.516a.953.953 0 1 0 0-1.906h-.058a3 3 0 0 0-.21.002 5 5 0 0 0-.53.044 3.5 3.5 0 0 0-.65.149c-.215.074-.431.182-.598.344-.26.254-.465.498-.598.825-.11.27-.164.581-.18.98Z"
        />
    </svg>
);
export default SvgGrocery;
