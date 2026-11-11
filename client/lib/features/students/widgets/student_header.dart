import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:phone/core/constants/app_radius.dart';
import 'package:phone/core/widgets/date_input_field.dart';
import 'package:phone/features/students/extensions/student_balance_extension.dart';
import 'package:phone/features/students/extensions/student_status_extension.dart';
import 'package:phone/features/students/widgets/student_hero_card.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:skeletonizer/skeletonizer.dart';

class StudentHeaderDelegate extends SliverPersistentHeaderDelegate {
  static const Color expandedColor = Color(0xFFCABEE4);
  static const double contentHeight = 136.0;

  final StudentSummaryResponse student;
  final List<TransactionResponse> payments;
  final double topPadding;
  final bool isRefreshing;
  final Widget? menu;
  final DateTime? now;
  final double? customContentHeight;

  const new({
    required this.student,
    required this.topPadding,
    this.payments = const [],
    this.isRefreshing = false,
    this.menu,
    this.now,
    this.customContentHeight,
  });

  static double computeContentHeight(
    StudentSummaryResponse student, {
    List<TransactionResponse> payments = const [],
    DateTime? now,
    BuildContext? context,
    double? screenWidth,
    TextScaler? textScaler,
  }) {
    final resolvedWidth =
        screenWidth ??
        (context != null ? MediaQuery.sizeOf(context).width : 360.0);
    final resolvedScaler =
        textScaler ??
        (context != null
            ? MediaQuery.textScalerOf(context)
            : TextScaler.noScaling);

    // Card padding: 16 top + 16 bottom = 32.0.
    // Horizontal space: screenWidth - 32 (card margins) - 32 (card padding) - 36 (avatar) - 12 (gap) = screenWidth - 112.0.
    final contentWidth = math.max(100.0, resolvedWidth - 112.0);
    final noteWidth = math.max(100.0, resolvedWidth - 64.0);

    const double avatarHeight = 36.0;

    // 1. Full name
    final namePainter = TextPainter(
      text: TextSpan(
        text: student.fullName,
        style: const TextStyle(
          fontSize: 16,
          fontWeight: FontWeight.w600,
          height: 1.5,
        ),
      ),
      textDirection: TextDirection.ltr,
      textScaler: resolvedScaler,
    )..layout(maxWidth: contentWidth);

    double columnHeight = namePainter.height;

    // 2. Groups
    final groups = student.studentGroups;
    if (groups != null && groups.isNotEmpty) {
      final validGroups = groups
          .where((g) => g.name.trim().isNotEmpty)
          .toList();
      if (validGroups.isNotEmpty) {
        columnHeight += 2.0; // SizedBox(height: 2)
        for (final group in validGroups) {
          final groupPainter = TextPainter(
            text: TextSpan(
              text: group.name,
              style: const TextStyle(fontSize: 14, height: 1.43),
            ),
            textDirection: TextDirection.ltr,
            textScaler: resolvedScaler,
          )..layout(maxWidth: contentWidth);
          columnHeight += groupPainter.height;
        }
      }
    }

    // 3. Birthdate
    if (student.birthdate != null) {
      columnHeight += 2.0;
      final datePainter = TextPainter(
        text: TextSpan(
          text: DateInputField.format(student.birthdate!),
          style: const TextStyle(fontSize: 14, height: 1.43),
        ),
        textDirection: TextDirection.ltr,
        textScaler: resolvedScaler,
      )..layout(maxWidth: contentWidth);
      columnHeight += datePainter.height;
    }

    // 4. Phone
    if (student.phone != null && student.phone!.isNotEmpty) {
      columnHeight += 2.0;
      final phonePainter = TextPainter(
        text: TextSpan(
          text: student.phone!,
          style: const TextStyle(fontSize: 14, height: 1.43),
        ),
        textDirection: TextDirection.ltr,
        textScaler: resolvedScaler,
      )..layout(maxWidth: contentWidth);
      columnHeight += phonePainter.height;
    }

    // 5. Badges
    final badges = [
      ...student.toHeaderStatusBadges(context),
      ...student.toBalanceBadges(context, payments, now),
    ];

    if (badges.isNotEmpty) {
      columnHeight += 8.0; // SizedBox(height: 8) before badges

      int lines = 1;
      double currentLineWidth = 0.0;

      for (final badge in badges) {
        final badgeLabelPainter = TextPainter(
          text: TextSpan(
            text: badge.label,
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
          ),
          textDirection: TextDirection.ltr,
          textScaler: resolvedScaler,
        )..layout();

        // 20.0 = horizontal padding (10 left + 10 right)
        // 22.0 = icon (16) + gap (6)
        final badgeWidth =
            20.0 + (badge.icon != null ? 22.0 : 0.0) + badgeLabelPainter.width;

        if (currentLineWidth == 0.0) {
          currentLineWidth = badgeWidth;
        } else if (currentLineWidth + 8.0 + badgeWidth <= contentWidth) {
          currentLineWidth += 8.0 + badgeWidth;
        } else {
          lines++;
          currentLineWidth = badgeWidth;
        }
      }

      final singleBadgeHeight = 10.0 + resolvedScaler.scale(20.0);
      const runSpacing = 8.0;
      final badgesHeight =
          (lines * singleBadgeHeight) + ((lines - 1) * runSpacing);
      columnHeight += badgesHeight;
    }

    final rowHeight = math.max(avatarHeight, columnHeight);
    double cardHeight = 32.0 + rowHeight + 5;

    // 6. Note
    final hasNote = student.note != null && student.note!.trim().isNotEmpty;
    if (hasNote) {
      cardHeight += 25.0; // 12 + 1 + 12
      final notePainter = TextPainter(
        text: TextSpan(
          text: student.note!,
          style: const TextStyle(fontSize: 14),
        ),
        textDirection: TextDirection.ltr,
        textScaler: resolvedScaler,
      )..layout(maxWidth: noteWidth);
      cardHeight += math.max(20.0, notePainter.height);
    }

    return cardHeight.ceilToDouble();
  }

  double get effectiveContentHeight =>
      customContentHeight ??
      computeContentHeight(student, payments: payments, now: now);

  @override
  double get minExtent => topPadding + kToolbarHeight;

  @override
  double get maxExtent => topPadding + kToolbarHeight + effectiveContentHeight;

  @override
  Widget build(
    BuildContext context,
    double shrinkOffset,
    bool overlapsContent,
  ) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final maxShrink = maxExtent - minExtent;
    final progress = maxShrink > 0
        ? (shrinkOffset / maxShrink).clamp(0.0, 1.0)
        : 0.0;

    // Smooth cross-fade between expanded and collapsed states
    final expandedOpacity = (1.0 - progress * 1.5).clamp(0.0, 1.0);
    final collapsedOpacity = ((progress - 0.35) / 0.65).clamp(0.0, 1.0);

    final curveProgress = Curves.easeInOut.transform(progress);
    final headerColor = Color.lerp(
      expandedColor,
      colorScheme.surface,
      curveProgress,
    )!;

    final badges = [
      ...student.toHeaderStatusBadges(context),
      ...student.toBalanceBadges(context, payments, now),
    ];

    const overlayStyle = SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      statusBarBrightness: Brightness.light,
    );

    // Purple header banner covers status bar, app bar, and top half of the card
    final cardHeight = effectiveContentHeight;
    final bannerOffset = math.min(136.0, cardHeight * 0.55);
    final currentBannerHeight = math.max(
      minExtent,
      (topPadding + kToolbarHeight + bannerOffset) - shrinkOffset,
    );
    final bannerBorderRadius = BorderRadius.lerp(
      BorderRadius.zero,
      AppRadius.bottomRounded,
      curveProgress,
    )!;

    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: overlayStyle,
      child: Stack(
        fit: StackFit.expand,
        children: [
          // 1. Purple Header Banner (covers status bar, app bar, and top half of the card)
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            height: currentBannerHeight,
            child: Material(
              color: headerColor,
              shape: RoundedRectangleBorder(borderRadius: bannerBorderRadius),
              clipBehavior: Clip.antiAlias,
              elevation: 0,
              child: const SizedBox.expand(),
            ),
          ),

          // 2. Student Card (half on header, half on body)
          if (expandedOpacity > 0)
            Positioned(
              top: (topPadding + kToolbarHeight) - shrinkOffset,
              left: 16,
              right: 16,
              child: Opacity(
                opacity: expandedOpacity,
                child: IgnorePointer(
                  ignoring: expandedOpacity < 0.2,
                  child: Skeletonizer(
                    enabled: isRefreshing,
                    child: StudentHeroCard(
                      student: student,
                      overrideBadges: badges,
                    ),
                  ),
                ),
              ),
            ),

          // 3. Collapsed / Pinned App Bar
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            height: minExtent,
            child: AppBar(
              systemOverlayStyle: overlayStyle,
              primary: true,
              backgroundColor: Colors.transparent,
              surfaceTintColor: Colors.transparent,
              elevation: 0,
              scrolledUnderElevation: 0,
              leading: const BackButton(),
              title: collapsedOpacity > 0
                  ? Opacity(
                      opacity: collapsedOpacity,
                      child: Text(
                        student.fullName,
                        style: const TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    )
                  : null,
              actions: [?menu, const SizedBox(width: 8)],
            ),
          ),
        ],
      ),
    );
  }

  @override
  bool shouldRebuild(covariant StudentHeaderDelegate oldDelegate) =>
      oldDelegate.student != student ||
      !listEquals(oldDelegate.payments, payments) ||
      oldDelegate.topPadding != topPadding ||
      oldDelegate.isRefreshing != isRefreshing ||
      oldDelegate.menu != menu ||
      oldDelegate.now != now ||
      oldDelegate.customContentHeight != customContentHeight;
}
