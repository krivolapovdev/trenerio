import 'package:flutter/material.dart';

class PaymentDateCard extends StatelessWidget {
  final String label;
  final String formattedDate;
  final bool canEdit;
  final VoidCallback onTap;
  final bool allowClear;
  final VoidCallback? onClear;
  final String? errorText;

  const new({
    super.key,
    required this.label,
    required this.formattedDate,
    required this.canEdit,
    required this.onTap,
    this.allowClear = false,
    this.onClear,
    this.errorText,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasError = errorText != null;

    return Material(
      color: Colors.white,
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: hasError
            ? BorderSide(color: theme.colorScheme.error)
            : BorderSide.none,
      ),
      child: InkWell(
        onTap: canEdit ? onTap : null,
        child: InputDecorator(
          decoration: InputDecoration(
            labelText: label,
            errorText: errorText,
            border: InputBorder.none,
            enabledBorder: InputBorder.none,
            focusedBorder: InputBorder.none,
            disabledBorder: InputBorder.none,
            contentPadding: const EdgeInsets.fromLTRB(16, 20, 16, 16),
            suffixIcon:
                allowClear &&
                    canEdit &&
                    formattedDate.isNotEmpty &&
                    onClear != null
                ? IconButton(
                    onPressed: onClear,
                    icon: const Icon(Icons.clear, size: 18),
                    padding: const EdgeInsets.all(8),
                    constraints: const BoxConstraints(),
                  )
                : null,
            suffixIconConstraints: const BoxConstraints(),
          ),
          child: Text(
            formattedDate.isEmpty ? '—' : formattedDate,
            style: TextStyle(
              fontSize: 16,
              color: formattedDate.isEmpty ? Colors.grey : null,
            ),
          ),
        ),
      ),
    );
  }
}
