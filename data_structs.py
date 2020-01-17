class MinHeapNode:

    def __init__(self, val, priority):
        self.val = val
        self.priority = priority


class MinHeap:
    def __init__(self):
        self.storage = []
        self.val_to_index = {}

    def _swap_elements(self, index_1, index_2):
        self.storage[index_1], self.storage[index_2] = self.storage[index_2], self.storage[index_1]
        for index in (index_1, index_2):
            self.val_to_index[self.storage[index].val] = index

    def get_priority(self, val):
        return self.storage[self.val_to_index[val]].priority

    def set(self, val, priority):
        index = self.val_to_index.get(val, None)
        if index is None:
            index = len(self.storage)
            self.storage.append(MinHeapNode(val, priority))
            self.val_to_index[self.storage[-1].val] = index
        else:
            self.storage[index].priority = priority

        self._balance_up(index)
        self._balance_down(index)

    def _balance_up(self, index):
        # while the node at index has a parent
        while self.parent_index(index) is not None:
            parent_index = self.parent_index(index)
            if self.storage[parent_index].priority <= self.storage[index].priority:
                return
            self._swap_elements(parent_index, index)
            index = parent_index

    def _balance_down(self, index):
        # while the node at index has children
        while self.left_child_index(index) is not None:
            index_to_swap = self.left_child_index(index)
            right_index = self.right_child_index(index)
            if right_index is not None and self.storage[right_index].priority < self.storage[index_to_swap].priority:
                index_to_swap = right_index

            if self.storage[index].priority <= self.storage[index_to_swap].priority:
                return
            self._swap_elements(index_to_swap, index)
            index = index_to_swap

    def pop(self):
        val = self.storage[0].val
        self.storage[0] = self.storage[-1]
        del self.val_to_index[val]
        del self.storage[-1]

        self._balance_down(0)
        return val

    @staticmethod
    def parent_index(i):
        if i == 0:
            return None
        return ((i + 1) // 2) - 1

    def left_child_index(self, i):
        index = 2 * (i + 1) - 1
        if index >= len(self.storage):
            return None
        return index

    def right_child_index(self, i):
        index = 2 * (i + 1)
        if index >= len(self.storage):
            return None
        return index

