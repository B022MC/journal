"use client";

import { roadmapData } from "@/data/roadmap";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { CheckCircle2, Circle, Flag } from "lucide-react";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { useEffect, useState } from "react";

export function ReactRoadmap() {
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
        setMounted(true);
    }, []);

    if (!mounted) {
        return null;
    }

    // Calculate overall progress
    const totalItems = roadmapData.reduce((acc, cat) => acc + cat.items.length, 0);
    const completedItems = roadmapData.reduce(
        (acc, cat) => acc + cat.items.filter((item) => item.completed).length,
        0
    );
    const progressPercentage = Math.round((completedItems / totalItems) * 100);

    return (
        <div className="container mx-auto py-12 px-4 space-y-12 max-w-5xl">
            <div className="space-y-4 text-center">
                <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl text-transparent bg-clip-text bg-gradient-to-r from-neutral-200 to-neutral-500">
                    S.H.I.T Journal Technical Roadmap
                </h1>
                <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
                    Social experiment addressing academic equality and content quality autonomy in an environment without administrative intervention or traditional peer review.
                </p>
            </div>

            <Card className="bg-neutral-900/50 border-neutral-800 shadow-2xl backdrop-blur-sm">
                <CardHeader>
                    <CardTitle className="flex items-center justify-between">
                        <span className="flex items-center gap-2">
                            <Flag className="w-5 h-5 text-neutral-400" />
                            Overall Progress
                        </span>
                        <span className="text-2xl font-bold">{progressPercentage}%</span>
                    </CardTitle>
                    <CardDescription>
                        {completedItems} of {totalItems} milestones achieved
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Progress value={progressPercentage} className="h-3" />
                </CardContent>
            </Card>

            <Separator className="bg-neutral-800" />

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                {roadmapData.map((category) => (
                    <div key={category.id} className="space-y-6">
                        <h2 className="text-2xl font-bold tracking-tight border-b border-neutral-800 pb-2">
                            {category.title}
                        </h2>
                        <div className="space-y-4">
                            {category.items.map((item) => (
                                <Card
                                    key={item.id}
                                    className={`border-neutral-800 bg-neutral-900/40 transition-all hover:bg-neutral-800/60 ${item.completed ? 'opacity-70' : ''}`}
                                >
                                    <CardHeader className="p-4 pl-5">
                                        <div className="flex items-start justify-between gap-4">
                                            <div className="flex items-start gap-3 flex-1">
                                                <div className="mt-0.5">
                                                    {item.completed ? (
                                                        <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                                                    ) : (
                                                        <Circle className="w-5 h-5 text-neutral-500" />
                                                    )}
                                                </div>
                                                <div className="space-y-1">
                                                    <CardTitle className={`text-base leading-tight ${item.completed ? "line-through text-neutral-400" : ""}`}>
                                                        {item.title}
                                                    </CardTitle>
                                                </div>
                                            </div>
                                            {item.priority && !item.completed && (
                                                <Badge
                                                    variant="outline"
                                                    className={`whitespace-nowrap ${item.priority.includes("High") ? "border-red-500/30 text-red-400 bg-red-500/10" :
                                                        item.priority.includes("Medium") ? "border-yellow-500/30 text-yellow-400 bg-yellow-500/10" :
                                                            "border-green-500/30 text-green-400 bg-green-500/10"
                                                        }`}
                                                >
                                                    {item.priority.split(" ")[1] === "High" ? "P0" : item.priority.split(" ")[1] === "Medium" ? "P1" : "P2"}
                                                </Badge>
                                            )}
                                        </div>
                                    </CardHeader>

                                    {item.details && item.details.length > 0 && (
                                        <CardContent className="p-0 border-t border-neutral-800/50">
                                            <Accordion className="w-full">
                                                <AccordionItem value={item.id} className="border-none">
                                                    <AccordionTrigger className="px-5 py-3 text-xs text-neutral-400 hover:text-neutral-300 hover:no-underline font-medium">
                                                        View details ({item.details.length})
                                                    </AccordionTrigger>
                                                    <AccordionContent className="px-5 pb-4 text-sm text-neutral-400">
                                                        <ul className="list-disc pl-4 space-y-1.5 marker:text-neutral-600">
                                                            {item.details.map((detail, idx) => {
                                                                const isHigh = detail.includes("🔴 High");
                                                                const isMed = detail.includes("🟡 Medium");
                                                                const isLow = detail.includes("🟢 Low");
                                                                let cleanDetail = detail;
                                                                if (isHigh || isMed || isLow) {
                                                                    cleanDetail = detail.replace(/^(🔴 High: |🟡 Medium: |🟢 Low: )/, '');
                                                                }

                                                                return (
                                                                    <li key={idx} className="leading-relaxed">
                                                                        {isHigh && <span className="inline-block w-2- h-2 rounded-full bg-red-500 mr-2 relative -top-0.5 shadow-[0_0_8px_rgba(239,68,68,0.5)]" />}
                                                                        {isMed && <span className="inline-block w-2- h-2 rounded-full bg-yellow-500 mr-2 relative -top-0.5 shadow-[0_0_8px_rgba(234,179,8,0.5)]" />}
                                                                        {isLow && <span className="inline-block w-2- h-2 rounded-full bg-green-500 mr-2 relative -top-0.5 shadow-[0_0_8px_rgba(34,197,94,0.5)]" />}
                                                                        {cleanDetail}
                                                                    </li>
                                                                );
                                                            })}
                                                        </ul>
                                                    </AccordionContent>
                                                </AccordionItem>
                                            </Accordion>
                                        </CardContent>
                                    )}
                                </Card>
                            ))}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
