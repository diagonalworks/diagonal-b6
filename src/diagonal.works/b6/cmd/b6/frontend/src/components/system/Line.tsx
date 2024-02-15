export interface LineProps extends React.HTMLAttributes<HTMLDivElement> {}

export const Line = ({ children }: LineProps) => {
    return (
        <div className="p-3 border max-w-80 min-h-11 border-graphite-30 hover:bg-graphite-10 ">
            {children}
        </div>
    );
};
