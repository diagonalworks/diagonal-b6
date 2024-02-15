import type { SVGProps } from 'react';
const SvgAlcoholShop = (props: SVGProps<SVGSVGElement>) => (
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
            d="M13.75 7.077v-.25h-3.577V9.77a1.79 1.79 0 0 0 1.154 1.666v2.314h-.135a.635.635 0 0 0 0 1.27h1.539a.635.635 0 0 0 0-1.27h-.135v-2.314A1.79 1.79 0 0 0 13.75 9.77V7.076Zm-8.577-1.34v.186c0 .205-.091.442-.263.718-.17.274-.4.557-.642.855l-.006.007c-.236.29-.485.596-.674.9s-.338.637-.338.982V14c0 .563.456 1.02 1.02 1.02h3.856l.011-.002c.53-.046.95-.466.997-.996V9.385c0-.331-.15-.662-.337-.965-.188-.304-.436-.617-.673-.915l-.005-.007a10 10 0 0 1-.644-.873q-.262-.421-.263-.702v-.186a.635.635 0 0 0 0-1.167v-.185a.635.635 0 0 0-.635-.635h-.77a.635.635 0 0 0-.634.635v.185a.635.635 0 0 0 0 1.167Zm7.308 4.032a.52.52 0 0 1-1.039 0V8.096h1.039zm-6.289 3.212a1.673 1.673 0 1 1 0-3.346 1.673 1.673 0 0 1 0 3.346Z"
        />
    </svg>
);
export default SvgAlcoholShop;
