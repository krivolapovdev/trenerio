import 'package:material_ui/material_ui.dart';

final BoxDecoration radialFabShadowDecoration = BoxDecoration(
  borderRadius: BorderRadius.circular(16),
  boxShadow: [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.1),
      blurRadius: 10,
      spreadRadius: 1,
    ),
  ],
);
