import { useTranslation } from 'react-i18next'
import { BarChart3 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts'
import { EmptyState } from '@/components/empty-state'

interface TopCategoriesChartProps {
  data: Array<{ category: string; count: number }>
}

export function TopCategoriesChart({ data }: TopCategoriesChartProps) {
  const { t } = useTranslation()

  const config = {
    count: { label: t('Detections'), color: 'hsl(var(--chart-1))' },
  }

  return (
    <Card className="flex flex-col">
      <CardHeader className="flex flex-row items-center gap-2">
        <BarChart3 className="size-4 text-muted-foreground" />
        <CardTitle>{t('Top Categories')}</CardTitle>
      </CardHeader>
      <CardContent className="flex-1">
        {data.length === 0 ? (
          <EmptyState
            icon={BarChart3}
            title={t('No Data')}
            description={t('No category detection data for the selected period.')}
            className="min-h-[200px] rounded-lg border border-dashed"
            bordered={false}
          />
        ) : (
          <ChartContainer config={config} className="min-h-[260px]">
            <BarChart data={data} margin={{ top: 8, right: 8, bottom: 24, left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis
                dataKey="category"
                angle={-30}
                textAnchor="end"
                height={60}
                tick={{ fontSize: 11 }}
                interval={0}
              />
              <YAxis allowDecimals={false} tick={{ fontSize: 11 }} />
              <ChartTooltip
                content={
                  <ChartTooltipContent
                    formatter={(value: any) => (
                      <span>
                        {t('Detections')}: {value}
                      </span>
                    )}
                  />
                }
              />
              <Bar dataKey="count" fill="hsl(var(--chart-1))" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
