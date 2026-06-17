/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { Tabs as TabsPrimitive } from '@base-ui/react/tabs'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'
import {
  createContext,
  useCallback,
  useContext,
  useLayoutEffect,
  useRef,
  useState,
} from 'react'

const TabsContext = createContext<{ value?: string }>({})

function Tabs({
  className,
  value,
  orientation = 'horizontal',
  ...props
}: TabsPrimitive.Root.Props) {
  return (
    <TabsContext.Provider value={{ value }}>
      <TabsPrimitive.Root
        data-slot='tabs'
        data-orientation={orientation}
        className={cn(
          'group/tabs flex gap-2 data-horizontal:flex-col',
          className
        )}
        {...(value !== undefined ? { value } : {})}
        {...props}
      />
    </TabsContext.Provider>
  )
}

const tabsListVariants = cva(
  'group/tabs-list inline-flex w-fit items-center justify-center rounded-lg p-[3px] text-muted-foreground group-data-horizontal/tabs:h-8 group-data-vertical/tabs:h-fit group-data-vertical/tabs:flex-col data-[variant=line]:rounded-none',
  {
    variants: {
      variant: {
        default: 'bg-muted',
        line: 'gap-1 bg-transparent',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

function TabsList({
  className,
  variant = 'default',
  ...props
}: TabsPrimitive.List.Props & VariantProps<typeof tabsListVariants>) {
  const { value } = useContext(TabsContext)
  const listRef = useRef<HTMLDivElement>(null)
  const [indicatorStyle, setIndicatorStyle] = useState({ left: 0, width: 0 })

  const updateIndicator = useCallback(() => {
    if (variant !== 'line') return
    // Base UI sets data-active as a present-but-empty attribute, not "true".
    const activeTrigger = listRef.current?.querySelector(
      '[data-active]'
    ) as HTMLElement | null
    if (!activeTrigger) return
    setIndicatorStyle({
      left: activeTrigger.offsetLeft,
      width: activeTrigger.offsetWidth,
    })
  }, [variant])

  useLayoutEffect(() => {
    if (variant !== 'line') return
    // Defer measurement to the next animation frame so Base UI has
    // finished updating the data-active attribute after value changes.
    const rafId = requestAnimationFrame(updateIndicator)
    return () => cancelAnimationFrame(rafId)
  }, [variant, value, updateIndicator])

  useLayoutEffect(() => {
    if (variant !== 'line') return
    const list = listRef.current
    if (!list) return

    const ro =
      typeof ResizeObserver !== 'undefined'
        ? new ResizeObserver(updateIndicator)
        : null
    ro?.observe(list)
    window.addEventListener('resize', updateIndicator)
    return () => {
      ro?.disconnect()
      window.removeEventListener('resize', updateIndicator)
    }
  }, [variant, updateIndicator])

  return (
    <div ref={listRef} className='relative'>
      <TabsPrimitive.List
        data-slot='tabs-list'
        data-variant={variant}
        className={cn(tabsListVariants({ variant }), className)}
        {...props}
      />
      {variant === 'line' && (
        <div
          className='pointer-events-none absolute bottom-[-5px] h-0.5 bg-foreground transition-all duration-300 ease-out will-change-[left,width]'
          style={{
            left: indicatorStyle.left,
            width: indicatorStyle.width,
          }}
        />
      )}
    </div>
  )
}

function TabsTrigger({ className, value, ...props }: TabsPrimitive.Tab.Props) {
  return (
    <TabsPrimitive.Tab
      data-slot='tabs-trigger'
      data-value={value}
      value={value}
      className={cn(
        "text-foreground/60 hover:text-foreground focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:outline-ring dark:text-muted-foreground dark:hover:text-foreground relative inline-flex h-[calc(100%-1px)] flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-1.5 py-0.5 text-sm font-medium whitespace-nowrap transition-all group-data-vertical/tabs:w-full group-data-vertical/tabs:justify-start focus-visible:ring-[3px] focus-visible:outline-1 disabled:pointer-events-none disabled:opacity-50 has-data-[icon=inline-end]:pr-1 has-data-[icon=inline-start]:pl-1 aria-disabled:pointer-events-none aria-disabled:opacity-50 group-data-[variant=default]/tabs-list:data-active:shadow-sm group-data-[variant=line]/tabs-list:data-active:shadow-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        'group-data-[variant=line]/tabs-list:bg-transparent group-data-[variant=line]/tabs-list:data-active:bg-transparent dark:group-data-[variant=line]/tabs-list:data-active:border-transparent dark:group-data-[variant=line]/tabs-list:data-active:bg-transparent',
        'data-active:bg-background data-active:text-foreground dark:data-active:border-input dark:data-active:bg-input/30 dark:data-active:text-foreground',
        className
      )}
      {...props}
    />
  )
}

function TabsContent({ className, ...props }: TabsPrimitive.Panel.Props) {
  return (
    <TabsPrimitive.Panel
      data-slot='tabs-content'
      className={cn('flex-1 text-sm outline-none', className)}
      {...props}
    />
  )
}

export { Tabs, TabsList, TabsTrigger, TabsContent, tabsListVariants }
