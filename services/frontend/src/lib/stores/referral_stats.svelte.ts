import { authFetch } from '$utils/authFetch';

interface NavigationPathItem {
	id: string;
	displayName: string;
}

interface CachedReferrals {
	referrals: DirectReferralStat[];
	pagination: PaginationMeta;
}

function createReferralStatsStore() {
	let navigationPath = $state<NavigationPathItem[]>([]);
	let loadedStats = $state<Map<string, UserReferralStats>>(new Map());
	let loadedReferrals = $state<Map<string, CachedReferrals>>(new Map());
	let currentStats = $state<UserReferralStats | null>(null);
	let currentReferrals = $state<DirectReferralStat[]>([]);
	let currentPagination = $state<PaginationMeta | null>(null);
	let isStatsLoading = $state<boolean>(false);
	let isReferralsLoading = $state<boolean>(false);
	let error = $state<string | null>(null);

	const getCurrentUserId = (): string | null => {
		if (navigationPath.length === 0) {
			return null; // Will use current user's ID from authStore
		}
		return navigationPath[navigationPath.length - 1].id;
	};

	const fetchStats = async (userId: string, showLoading: boolean = true) => {
		// Check cache first
		const cached = loadedStats.get(userId);
		if (cached) {
			return cached;
		}

		if (showLoading) {
			isStatsLoading = true;
		}
		error = null;

		const resp = await authFetch(`/desim/stats/${userId}`);
		if (!resp.ok) {
			error = 'Failed to fetch stats';
			if (showLoading) {
				isStatsLoading = false;
			}
			return null;
		}

		const response: ApiResponse<UserReferralStats> = await resp.json();
		if (resp.status !== 200 || !response.data) {
			error = response.error || 'Failed to fetch stats';
			if (showLoading) {
				isStatsLoading = false;
			}
			return null;
		}

		loadedStats.set(userId, response.data);
		if (showLoading) {
			isStatsLoading = false;
		}
		return response.data;
	};

	const fetchFirstLine = async (userId: string, page: number = 1, take: number = 10, searchTerm: string = '') => {
		const cacheKey = `${userId}_${page}_${take}_${searchTerm || ''}`;
		const cached = loadedReferrals.get(cacheKey);
		if (cached) {
			return cached;
		}

		isReferralsLoading = true;
		error = null;

		const params = new URLSearchParams({
			page: page.toString(),
			take: take.toString(),
		});
		if (searchTerm) {
			params.append('q', searchTerm);
		}

		const resp = await authFetch(`/desim/stats/${userId}/firstLine?${params}`);
		if (!resp.ok) {
			error = 'Failed to fetch referrals';
			isReferralsLoading = false;
			return null;
		}

		const response: ApiPaginatedResponse<DirectReferralStat> = await resp.json();
		if (resp.status !== 200) {
			error = response.error || 'Failed to fetch referrals';
			isReferralsLoading = false;
			return null;
		}

		const result: CachedReferrals = {
			referrals: response.data || [],
			pagination: response.meta || {
				page: 1,
				take: 10,
				itemCount: 0,
				pageCount: 0,
			},
		};

		loadedReferrals.set(cacheKey, result);
		isReferralsLoading = false;
		return result;
	};

	const updateCurrentView = async () => {
		const userId = getCurrentUserId();
		if (!userId) {
			// Root view - will be handled by page component with authStore.user.id
			currentStats = null;
			currentReferrals = [];
			currentPagination = null;
			return;
		}

		const stats = loadedStats.get(userId);
		if (stats) {
			currentStats = stats;
		} else {
			const fetchedStats = await fetchStats(userId);
			if (fetchedStats) {
				currentStats = fetchedStats;
			}
		}

		const cached = loadedReferrals.get(`${userId}_1_10_`);
		if (cached) {
			currentReferrals = cached.referrals;
			currentPagination = cached.pagination;
		} else {
			const fetched = await fetchFirstLine(userId, 1, 10, '');
			if (fetched) {
				currentReferrals = fetched.referrals;
				currentPagination = fetched.pagination;
			}
		}
	};

	return {
		get navigationPath() {
			return navigationPath;
		},
		get currentStats() {
			return currentStats;
		},
		get currentReferrals() {
			return currentReferrals;
		},
		get currentPagination() {
			return currentPagination;
		},
		get isStatsLoading() {
			return isStatsLoading;
		},
		get isReferralsLoading() {
			return isReferralsLoading;
		},
		get error() {
			return error;
		},
		set error(value) {
			error = value;
		},
		fetchStats,
		fetchFirstLine,
		async navigateTo(userId: string, displayName: string, stats?: UserReferralStats) {
			// Update navigation path first (immediate UI feedback)
			navigationPath = [...navigationPath, { id: userId, displayName }];

			// Use provided stats if available (from clicked referral), otherwise fetch
			if (stats) {
				currentStats = stats;
				loadedStats.set(userId, stats);
			} else {
				// Only fetch stats if not cached (don't show loading if we have cached data)
				const cachedStats = loadedStats.get(userId);
				if (cachedStats) {
					currentStats = cachedStats;
				} else {
					const statsResp = await fetchStats(userId, true);
					if (statsResp) {
						currentStats = statsResp;
					}
				}
			}

			// Fetch referrals (this will show loading state)
			const referralsResp = await fetchFirstLine(userId);
			if (referralsResp) {
				currentReferrals = referralsResp.referrals;
				currentPagination = referralsResp.pagination;
			}
		},
		async navigateUp(rootUserId?: string) {
			if (navigationPath.length > 0) {
				const willBeAtRoot = navigationPath.length === 1;
				navigationPath = navigationPath.slice(0, -1);
				
				if (willBeAtRoot && rootUserId) {
					// Load root view
					navigationPath = [];
					const statsResp = await fetchStats(rootUserId);
					const referralsResp = await fetchFirstLine(rootUserId);

					if (statsResp) {
						currentStats = statsResp;
					}

					if (referralsResp) {
						currentReferrals = referralsResp.referrals;
						currentPagination = referralsResp.pagination;
					}
				} else {
					await updateCurrentView();
				}
			}
		},
		async navigateToRoot(rootUserId?: string) {
			navigationPath = [];
			if (rootUserId) {
				const statsResp = await fetchStats(rootUserId);
				const referralsResp = await fetchFirstLine(rootUserId);

				if (statsResp) {
					currentStats = statsResp;
				}

				if (referralsResp) {
					currentReferrals = referralsResp.referrals;
					currentPagination = referralsResp.pagination;
				}
			} else {
				await updateCurrentView();
			}
		},
		async navigateToPath(index: number, rootUserId?: string) {
			if (index === -1 && rootUserId) {
				// Clicking "My Stats" breadcrumb - go to root
				navigationPath = [];
				const statsResp = await fetchStats(rootUserId);
				const referralsResp = await fetchFirstLine(rootUserId);

				if (statsResp) {
					currentStats = statsResp;
				}

				if (referralsResp) {
					currentReferrals = referralsResp.referrals;
					currentPagination = referralsResp.pagination;
				}
			} else if (index >= 0 && index < navigationPath.length) {
				const willBeAtRoot = index === 0 && navigationPath.length === 1;
				navigationPath = navigationPath.slice(0, index + 1);
				
				if (willBeAtRoot && rootUserId) {
					// Load root view
					navigationPath = [];
					const statsResp = await fetchStats(rootUserId);
					const referralsResp = await fetchFirstLine(rootUserId);

					if (statsResp) {
						currentStats = statsResp;
					}

					if (referralsResp) {
						currentReferrals = referralsResp.referrals;
						currentPagination = referralsResp.pagination;
					}
				} else {
					await updateCurrentView();
				}
			}
		},
		updateCurrentView,
		async loadRootView(userId: string) {
			navigationPath = [];
			const statsResp = await fetchStats(userId);
			const referralsResp = await fetchFirstLine(userId);

			if (statsResp) {
				currentStats = statsResp;
			}

			if (referralsResp) {
				currentReferrals = referralsResp.referrals;
				currentPagination = referralsResp.pagination;
			}
		},
		async refreshCurrentView(page: number = 1, take: number = 10, searchTerm: string = '', userId?: string) {
			const targetUserId = userId || getCurrentUserId();
			if (!targetUserId) {
				return;
			}

			// Only fetch stats if not cached (avoid unnecessary loading state)
			const cachedStats = loadedStats.get(targetUserId);
			if (!cachedStats) {
				const statsResp = await fetchStats(targetUserId, false);
				if (statsResp) {
					currentStats = statsResp;
					loadedStats.set(targetUserId, statsResp);
				}
			} else {
				// Ensure current stats is set even if cached
				currentStats = cachedStats;
			}

			// Always fetch referrals (they change with page/search)
			const referralsResp = await fetchFirstLine(targetUserId, page, take, searchTerm);
			if (referralsResp) {
				currentReferrals = referralsResp.referrals;
				currentPagination = referralsResp.pagination;
				loadedReferrals.set(`${targetUserId}_${page}_${take}_${searchTerm || ''}`, referralsResp);
			}
		},
		clear() {
			navigationPath = [];
			loadedStats.clear();
			loadedReferrals.clear();
			currentStats = null;
			currentReferrals = [];
			currentPagination = null;
			error = null;
		},
	};
}

export const referralStatsStore = createReferralStatsStore();
