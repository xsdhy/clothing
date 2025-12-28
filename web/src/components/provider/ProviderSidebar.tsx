import React from "react";
import {
    Box,
    Button,
    List,
    ListItemButton,
    Typography,
    Chip,
    Stack,
    Paper,
    InputBase,
} from "@mui/material";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import type { ProviderAdmin } from "../../types";

interface ProviderSidebarProps {
    providers: ProviderAdmin[];
    selectedProviderId: string | null;
    onSelect: (providerId: string) => void;
    onAddProvider: () => void;
}

const ProviderSidebar: React.FC<ProviderSidebarProps> = ({
    providers,
    selectedProviderId,
    onSelect,
    onAddProvider,
}) => {
    return (
        <Paper
            elevation={0}
            sx={{
                width: { xs: "100%", md: 240 }, // Further reduced width
                height: "100%",
                display: "flex",
                flexDirection: "column",
                borderRight: "1px solid",
                borderColor: "divider",
                borderRadius: { xs: 2, md: "12px 0 0 12px" },
                overflow: "hidden",
                bgcolor: "background.paper",
            }}
        >
            <Box sx={{ p: 1.5, borderBottom: "1px solid", borderColor: "divider" }}>
                <Stack
                    direction="row"
                    alignItems="center"
                    justifyContent="space-between"
                    sx={{ mb: 1.5 }}
                >
                    <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                        厂商列表
                    </Typography>
                    <Button
                        variant="contained"
                        size="small"
                        startIcon={<AddRoundedIcon sx={{ fontSize: 18 }} />}
                        onClick={onAddProvider}
                        sx={{ minWidth: 64, px: 1, py: 0.5, fontSize: "0.8rem" }}
                    >
                        新增
                    </Button>
                </Stack>
                <Paper
                    elevation={0}
                    sx={{
                        p: "2px 8px",
                        display: "flex",
                        alignItems: "center",
                        width: "100%",
                        bgcolor: "action.hover",
                        borderRadius: 1.5,
                        height: 32,
                    }}
                >
                    <SearchRoundedIcon sx={{ color: "text.secondary", fontSize: 18 }} />
                    <InputBase
                        sx={{ ml: 1, flex: 1, fontSize: "0.85rem" }}
                        placeholder="搜索厂商..."
                    />
                </Paper>
            </Box>

            <List sx={{ overflowY: "auto", flex: 1, p: 1 }}>
                {providers.length > 0 ? (
                    providers.map((provider) => (
                        <ListItemButton
                            key={provider.id}
                            selected={selectedProviderId === provider.id}
                            onClick={() => onSelect(provider.id)}
                            sx={{
                                borderRadius: 1.5,
                                mb: 0.5,
                                py: 0.75,
                                px: 1.5,
                                "&.Mui-selected": {
                                    bgcolor: "primary.light",
                                    color: "primary.contrastText",
                                    "&:hover": {
                                        bgcolor: "primary.main",
                                    },
                                    "& .MuiTypography-root": {
                                        color: "inherit",
                                    },
                                    "& .MuiChip-root": {
                                        bgcolor: "rgba(255,255,255,0.2)",
                                        color: "inherit",
                                        borderColor: "transparent",
                                    },
                                },
                            }}
                        >
                            <Box sx={{ width: "100%" }}>
                                <Stack
                                    direction="row"
                                    justifyContent="space-between"
                                    alignItems="center"
                                    sx={{ mb: 0.25 }}
                                >
                                    <Typography
                                        variant="body2"
                                        sx={{ fontWeight: 600, color: "text.primary" }}
                                    >
                                        {provider.name}
                                    </Typography>
                                    {!provider.is_active && (
                                        <Chip
                                            label="停用"
                                            size="small"
                                            color="error"
                                            variant="outlined"
                                            sx={{ height: 16, fontSize: "0.65rem", border: "none", bgcolor: "error.soft", px: 0 }}
                                        />
                                    )}
                                </Stack>
                                <Stack
                                    direction="row"
                                    alignItems="center"
                                    justifyContent="space-between"
                                >
                                    <Typography
                                        variant="caption"
                                        color="text.secondary"
                                        sx={{ display: "flex", alignItems: "center", gap: 0.5, fontSize: "0.75rem" }}
                                    >
                                        {provider.driver} · {provider.models?.length ?? 0}
                                    </Typography>
                                </Stack>
                            </Box>
                        </ListItemButton>
                    ))
                ) : (
                    <Box sx={{ p: 3, textAlign: "center" }}>
                        <Typography variant="body2" color="text.secondary">
                            暂无厂商
                        </Typography>
                    </Box>
                )}
            </List>
        </Paper>
    );
};

export default ProviderSidebar;
