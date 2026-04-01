let count = $state(0);

export const unreadStore = {
	get count() {
		return count;
	},
	increment() {
		count++;
	},
	reset() {
		count = 0;
	}
};
