permit(
	principal in Role::"Triager",
	action in [Action::"close", Action::"label", Action::"remove-label"],
	resource is Issue
);
