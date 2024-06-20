import * as Collapsible from '@radix-ui/react-collapsible';
import { TriangleDownIcon, TriangleUpIcon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { ErrorBoundaryPropsWithComponent } from 'react-error-boundary';

import { Button } from '@/components/system/Button';

export const MainErrorFallback: ErrorBoundaryPropsWithComponent['FallbackComponent'] =
    ({ error, resetErrorBoundary }) => {
        const [open, setOpen] = useState(false);
        return (
            <div className=" bg-graphite-20  min-h-screen w-screen flex flex-col pt-32">
                <div
                    className=" w-1/2 flex flex-col item-center mx-auto "
                    role="alert"
                >
                    <h2 className="text-2xl font-medium text-center">
                        Oh no! It seems like you've stumbled upon uncharted
                        territory.
                    </h2>

                    <Button
                        className="mt-4 mx-auto "
                        onClick={resetErrorBoundary}
                    >
                        Refresh
                    </Button>
                    <Collapsible.Root open={open} onOpenChange={setOpen}>
                        <Collapsible.Trigger className=" text-left border-b border-amber-70 text-amber-70 w-full ">
                            <button className="mt-4 flex gap-2 items-center">
                                {open ? (
                                    <TriangleUpIcon />
                                ) : (
                                    <TriangleDownIcon />
                                )}

                                {error.message}
                            </button>
                        </Collapsible.Trigger>
                        <Collapsible.Content className=" bg-amber-20/45 px-4 py-2 text-amber-90 text-xs  overflow-auto">
                            <pre>{error.stack}</pre>
                        </Collapsible.Content>
                    </Collapsible.Root>
                </div>
            </div>
        );
    };
