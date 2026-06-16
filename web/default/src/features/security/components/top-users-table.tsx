import { Users } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { EmptyState } from '@/components/empty-state'

interface TopUsersTableProps {
  data: Array<{ user_id: number; user_name: string; count: number }>
}

export function TopUsersTable({ data }: TopUsersTableProps) {
  const { t } = useTranslation()

  return (
    <Card className="flex flex-col">
      <CardHeader className="flex flex-row items-center gap-2">
        <Users className="size-4 text-muted-foreground" />
        <CardTitle>{t('Top Users')}</CardTitle>
      </CardHeader>
      <CardContent className="flex-1 p-0">
        {data.length === 0 ? (
          <EmptyState
            icon={Users}
            title={t('No Data')}
            description={t('No user detection data for the selected period.')}
            className="min-h-[180px] rounded-lg border border-dashed"
            bordered={false}
          />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-16">#</TableHead>
                <TableHead>{t('User')}</TableHead>
                <TableHead className="text-right">{t('Detections')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((item, idx) => (
                <TableRow key={item.user_id ?? idx}>
                  <TableCell className="font-medium">{idx + 1}</TableCell>
                  <TableCell>{item.user_name || `User #${item.user_id}`}</TableCell>
                  <TableCell className="text-right">{item.count}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
