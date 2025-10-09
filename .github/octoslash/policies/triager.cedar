permit(
	principal in Role::"Triager",
	action in [Action::"close", Action::"add-label", Action::"remove-label", Action::"self-assign", Action::"self-unassign"],
	resource is Issue
);
