import 'package:material_ui/material_ui.dart';
import 'package:phone/core/constants/app_colors.dart';

abstract final class AppTheme {
  static final ThemeData light = ThemeData(
    useMaterial3: true,
    scaffoldBackgroundColor: AppColors.background,
  );
}
