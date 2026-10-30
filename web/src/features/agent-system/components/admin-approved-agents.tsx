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

import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatQuota } from "@/lib/format";

import { listApprovedAgents } from "../api";
import type { ApprovedAgent } from "../types";

function rateLabel(rate: number, custom: boolean, customLabel: string) {
  return custom ? `${rate}% (${customLabel})` : `${rate}%`;
}

export function AdminApprovedAgents() {
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const query = useQuery({
    queryKey: ["approved-agents", page],
    queryFn: () => listApprovedAgents(page),
  });
  const items = query.data?.items ?? [];
  const total = query.data?.total ?? 0;

  return (
    <section className="bg-card mb-4 space-y-4 rounded-xl border p-4 sm:p-5">
      <div>
        <h3 className="font-semibold">{t("Approved agents")}</h3>
        <p className="text-muted-foreground mt-1 text-sm">
          {t(
            "Every approved agent is listed here. Search below opens one user for settlement or a custom rate.",
          )}
        </p>
      </div>
      {query.isError ? (
        <p className="text-sm text-red-600 dark:text-red-400">
          {t("Failed to load approved agents")}
        </p>
      ) : null}
      {query.isLoading ? (
        <p className="text-muted-foreground text-sm">{t("Loading...")}</p>
      ) : null}
      {!query.isLoading && !query.isError && items.length === 0 ? (
        <p className="text-muted-foreground text-sm">
          {t("No approved agents")}
        </p>
      ) : null}
      {items.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("User")}</TableHead>
              <TableHead>{t("Invite code")}</TableHead>
              <TableHead>{t("Account quota")}</TableHead>
              <TableHead>{t("Used quota")}</TableHead>
              <TableHead>{t("Available commission")}</TableHead>
              <TableHead>{t("Total commission")}</TableHead>
              <TableHead>{t("Transferred to balance")}</TableHead>
              <TableHead>{t("Settled by support")}</TableHead>
              <TableHead>{t("Level 1 users")}</TableHead>
              <TableHead>{t("Level 1 commission")}</TableHead>
              <TableHead>{t("Level 1 rate")}</TableHead>
              <TableHead>{t("Level 2 users")}</TableHead>
              <TableHead>{t("Level 2 commission")}</TableHead>
              <TableHead>{t("Level 2 rate")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <AgentRow key={item.id} item={item} />
            ))}
          </TableBody>
        </Table>
      ) : null}
      {total > 20 ? (
        <div className="flex items-center justify-between gap-3">
          <p className="text-muted-foreground text-sm">
            {t("{{count}} records", { count: total })}
          </p>
          <div className="flex gap-2">
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={page <= 1}
              onClick={() => setPage((current) => current - 1)}
            >
              {t("Previous")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={page * 20 >= total}
              onClick={() => setPage((current) => current + 1)}
            >
              {t("Next")}
            </Button>
          </div>
        </div>
      ) : null}
    </section>
  );
}

function AgentRow({ item }: { item: ApprovedAgent }) {
  const { t } = useTranslation();
  return (
    <TableRow>
      <TableCell>
        <div className="whitespace-nowrap">
          {item.displayName || item.username}
        </div>
        <div
          data-table-text="secondary"
          className="text-muted-foreground whitespace-nowrap"
        >
          {item.username} · #{item.id}
        </div>
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {item.inviteCode || "—"}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.quota)}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.usedQuota)}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.availableQuota)}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.totalQuota)}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.transferredQuota)}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.deductedQuota)}
      </TableCell>
      <TableCell>{item.level1Count}</TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.level1CommissionQuota)}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {rateLabel(item.level1Rate, item.level1Custom, t("Custom rate"))}
      </TableCell>
      <TableCell>{item.level2Count}</TableCell>
      <TableCell className="whitespace-nowrap">
        {formatQuota(item.level2CommissionQuota)}
      </TableCell>
      <TableCell className="whitespace-nowrap">
        {rateLabel(item.level2Rate, item.level2Custom, t("Custom rate"))}
      </TableCell>
    </TableRow>
  );
}
